use crate::models::*;
use futures_util::future;
use reqwest::blocking::Client;
use scraper::Selector;
use thiserror::Error;

#[derive(Debug, Error)]
pub enum ApiError {
    #[error(transparent)]
    NetworkError(#[from] reqwest::Error),

    // #[error("cannot parse the response")]
    #[error(transparent)]
    DeserializeError(#[from] serde_json::Error),

    #[error("cannot find any data")]
    BundleNotFound,
}

pub struct HumbleApi {
    auth_key: String,
}

impl HumbleApi {
    pub fn new(auth_key: &str) -> Self {
        Self {
            auth_key: auth_key.to_owned(),
        }
    }

    pub fn list_bundle_keys(&self) -> Result<Vec<String>, ApiError> {
        let client = Client::new();

        let res = client
            .get("https://www.humblebundle.com/api/v1/user/order")
            .header(reqwest::header::ACCEPT, "application/json")
            .header(
                "cookie".to_owned(),
                format!("_simpleauth_sess={}", self.auth_key),
            )
            .send()?
            .error_for_status()?;

        let game_keys = res
            .json::<Vec<GameKey>>()?
            .into_iter()
            .map(|g| g.gamekey)
            .collect();

        Ok(game_keys)
    }

    pub fn list_bundles(&self) -> Result<Vec<Bundle>, ApiError> {
        const CHUNK_SIZE: usize = 10;

        let client = reqwest::Client::new();
        let game_keys = self.list_bundle_keys()?;

        let runtime = tokio::runtime::Builder::new_multi_thread()
            .enable_all()
            .build()
            .expect("cannot build the tokio runtime");

        let futures = game_keys
            .chunks(CHUNK_SIZE)
            .map(|keys| self.read_bundles_data(&client, keys));

        // Collect the Vec<Result<_,_>> into Result<Vec<_>, _>. This will automatically stop when an error is seen.
        // See https://doc.rust-lang.org/rust-by-example/error/iter_result.html#fail-the-entire-operation-with-collect
        let result: Result<Vec<Vec<Bundle>>, _> = runtime
            .block_on(future::join_all(futures))
            .into_iter()
            .collect();

        let mut bundles: Vec<_> = result?.into_iter().flatten().collect();
        bundles.sort_by(|a, b| a.created.partial_cmp(&b.created).unwrap());
        Ok(bundles)
    }

    async fn read_bundles_data(
        &self,
        client: &reqwest::Client,
        keys: &[String],
    ) -> Result<Vec<Bundle>, ApiError> {
        let mut query_params: Vec<_> = keys.iter().map(|key| ("gamekeys", key.as_str())).collect();

        query_params.insert(0, ("all_tpkds", "true"));

        let res = client
            .get("https://www.humblebundle.com/api/v1/orders")
            .header(reqwest::header::ACCEPT, "application/json")
            .header(
                "cookie".to_owned(),
                format!("_simpleauth_sess={}", self.auth_key),
            )
            .query(&query_params)
            .send()
            .await?
            .error_for_status()?;

        let product_map = res.json::<BundleMap>().await?;
        Ok(product_map.into_values().collect())
    }

    pub fn read_bundle(&self, product_key: &str) -> Result<Bundle, ApiError> {
        let url = format!(
            "https://www.humblebundle.com/api/v1/order/{}?all_tpkds=true",
            product_key
        );

        let client = Client::new();
        let res = client
            .get(url)
            .header(reqwest::header::ACCEPT, "application/json")
            .header(
                "cookie".to_owned(),
                format!("_simpleauth_sess={}", self.auth_key),
            )
            .send()?
            .error_for_status()?;

        res.json::<Bundle>().map_err(|e| e.into())
    }

    /// Read Bundle Choices for the give month and year.
    ///
    /// `when` should be in the `month-year` format. For example: `"january-2023"`.
    /// Use `"home"` to get the current active data.
    pub fn read_bundle_choices(&self, when: &str) -> Result<HumbleChoice, ApiError> {
        let url = format!("https://www.humblebundle.com/membership/{}", when);

        let client = Client::new();
        let res = client
            .get(url)
            .header(
                "cookie".to_owned(),
                format!("_simpleauth_sess={}", self.auth_key),
            )
            .send()?
            .error_for_status()?;

        let html = res.text()?;
        self.parse_bundle_choices(&html)
    }

    fn parse_bundle_choices(&self, html: &str) -> Result<HumbleChoice, ApiError> {
        let document = scraper::html::Html::parse_document(html);
        // One of these two CSS IDs will match. First one is for the active
        // month, while the second is for previous months.
        let sel = Selector::parse(
            "script#webpack-subscriber-hub-data, script#webpack-monthly-product-data",
        )
        .unwrap();

        let scripts: Vec<_> = document.select(&sel).collect();
        if scripts.len() != 1 {
            return Err(ApiError::BundleNotFound);
        }

        let script = scripts.get(0).unwrap();
        let txt = script.inner_html();
        let obj: HumbleChoice = serde_json::from_str(&txt)?;
        Ok(obj)
    }
}
