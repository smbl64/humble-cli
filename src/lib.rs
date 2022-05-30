use reqwest::Client;
use serde::Deserialize;
use std::collections::HashMap;
use thiserror::Error;

#[derive(Debug, Error)]
pub enum ApiError {
    #[error("bad status code: {0}")]
    BadHttpStatus(u16),
}

type ProductMap = HashMap<String, Product>;

#[derive(Debug, Deserialize)]
pub struct Product {
    pub gamekey: String,
    #[serde(rename = "subproducts")]
    pub entries: Vec<ProductEntry>,
}

impl Product {
    pub fn total_size(&self) -> u64 {
        self.entries.iter().map(|e| e.total_size()).sum()
    }
}

#[derive(Debug, Deserialize)]
pub struct ProductEntry {
    machine_name: String,
    human_name: String,
    #[serde(rename = "url")]
    product_details_url: String,
    downloads: Vec<DownloadEntry>,
}

impl ProductEntry {
    fn total_size(&self) -> u64 {
        self.downloads.iter().map(|e| e.total_size()).sum()
    }
}

#[derive(Debug, Deserialize)]
struct DownloadEntry {
    #[serde(rename = "download_struct")]
    sub_items: Vec<DownloadEntryItem>,
}

impl DownloadEntry {
    pub fn total_size(&self) -> u64 {
        self.sub_items.iter().map(|e| e.file_size).sum()
    }
}

#[derive(Debug, Deserialize)]
struct DownloadEntryItem {
    md5: String,
    #[serde(rename = "name")]
    item_type: String,
    file_size: u64,

    #[serde(rename = "url")]
    urls: DownloadUrl,
}

#[derive(Debug, Deserialize)]
struct DownloadUrl {
    web: String,
    bittorrent: String,
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

    pub async fn list_products(&self) -> Result<Vec<Product>, anyhow::Error> {
        let client = Client::new();

        // First: get the game keys
        let res = client
            .get("https://www.humblebundle.com/api/v1/user/order")
            .header(reqwest::header::ACCEPT, "application/json")
            .header(
                "cookie".to_owned(),
                format!("_simpleauth_sess={}", self.auth_key),
            )
            .send()
            .await?;

        if !res.status().is_success() {
            return Err(ApiError::BadHttpStatus(res.status().as_u16()).into());
        }

        let game_keys = res.json::<Vec<GameKey>>().await?;
        let game_keys: Vec<_> = game_keys.into_iter().map(|g| g.gamekey).collect();

        // Second: get details for those game keys
        let query_params: Vec<_> = game_keys.into_iter().map(|key| ("gamekeys", key)).collect();

        let res = client
            .get("https://www.humblebundle.com/api/v1/orders")
            .header(reqwest::header::ACCEPT, "application/json")
            .header(
                "cookie".to_owned(),
                format!("_simpleauth_sess={}", self.auth_key),
            )
            .query(&query_params)
            .send()
            .await?;

        if !res.status().is_success() {
            return Err(ApiError::BadHttpStatus(res.status().as_u16()).into());
        }
        let product_map = res.json::<ProductMap>().await?;
        Ok(product_map.into_values().collect())
    }

    pub async fn read_product(&self, product_key: &str) -> Result<Product, anyhow::Error> {
        let url = format!("https://www.humblebundle.com/api/v1/order/{}", product_key);

        let client = Client::new();
        let res = client
            .get(url)
            .header(reqwest::header::ACCEPT, "application/json")
            .header(
                "cookie".to_owned(),
                format!("_simpleauth_sess={}", self.auth_key),
            )
            .send()
            .await?;

        if !res.status().is_success() {
            //eprintln!("Request failed: {}", res.status().as_u16());
            return Err(ApiError::BadHttpStatus(res.status().as_u16()).into());
        }
        let product = res.json::<Product>().await?;

        //println!("{:?}", product);

        //for e in product.entries {
        //    println!("{}", e.human_name);
        //    for d in e.downloads {
        //        for sub_item in d.sub_items {
        //            println!(" {}: {}", sub_item.item_type, sub_item.urls.web);
        //        }
        //    }
        //}

        Ok(product)
    }
}
