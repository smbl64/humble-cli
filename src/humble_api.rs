use futures_util::future;
use reqwest::blocking::Client;
use serde::Deserialize;
use serde_with::{serde_as, VecSkipError};
use std::collections::HashMap;
use thiserror::Error;

#[derive(Debug, Error)]
pub enum ApiError {
    #[error(transparent)]
    NetworkError(#[from] reqwest::Error),

    #[error("Cannot parse the response")]
    DeserializeFailed,
}

type BundleMap = HashMap<String, Bundle>;

#[serde_as]
#[derive(Debug, Deserialize)]
pub struct Bundle {
    pub gamekey: String,

    #[serde(rename = "product")]
    pub details: BundleDetails,

    #[serde(rename = "subproducts")]
    #[serde_as(as = "VecSkipError<_>")]
    pub products: Vec<Product>,
}

#[derive(Debug, Deserialize)]
pub struct BundleDetails {
    pub machine_name: String,
    pub human_name: String,
}

impl Bundle {
    pub fn total_size(&self) -> u64 {
        self.products.iter().map(|e| e.total_size()).sum()
    }
}

#[derive(Debug, Deserialize)]
pub struct Product {
    pub machine_name: String,
    pub human_name: String,

    #[serde(rename = "url")]
    pub product_details_url: String,

    /// List of associated downloads with this product.
    ///
    /// Note: Each product usually has one item here.
    pub downloads: Vec<ProductDownload>,
}

impl Product {
    pub fn total_size(&self) -> u64 {
        self.downloads.iter().map(|e| e.total_size()).sum()
    }

    pub fn formats_as_vec(&self) -> Vec<&str> {
        self.downloads
            .iter()
            .flat_map(|d| d.formats_as_vec())
            .collect::<Vec<_>>()
    }

    pub fn formats(&self) -> String {
        self.formats_as_vec().join(", ")
    }
}

#[derive(Debug, Deserialize)]
pub struct ProductDownload {
    #[serde(rename = "download_struct")]
    pub items: Vec<DownloadInfo>,
}

impl ProductDownload {
    pub fn total_size(&self) -> u64 {
        self.items.iter().map(|e| e.file_size).sum()
    }

    pub fn formats_as_vec(&self) -> Vec<&str> {
        self.items.iter().map(|s| &s.format[..]).collect::<Vec<_>>()
    }

    pub fn formats(&self) -> String {
        self.formats_as_vec().join(", ")
    }
}

#[derive(Debug, Deserialize)]
pub struct DownloadInfo {
    pub md5: String,

    #[serde(rename = "name")]
    pub format: String,

    pub file_size: u64,

    pub url: DownloadUrl,
}

#[derive(Debug, Deserialize)]
pub struct DownloadUrl {
    pub web: String,
    pub bittorrent: String,
}

#[derive(Debug, Deserialize)]
struct GameKey {
    gamekey: String,
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

        Ok(result?.into_iter().flatten().collect())
    }

    async fn read_bundles_data(
        &self,
        client: &reqwest::Client,
        keys: &[String],
    ) -> Result<Vec<Bundle>, ApiError> {
        let query_params: Vec<_> = keys.into_iter().map(|key| ("gamekeys", key)).collect();

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
        let url = format!("https://www.humblebundle.com/api/v1/order/{}", product_key);

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

        res.json::<Bundle>()
            .map_err(|_| ApiError::DeserializeFailed)
    }
}
