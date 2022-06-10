use reqwest::blocking::Client;
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

    #[serde(rename = "product")]
    pub details: ProductDetails,

    #[serde(rename = "subproducts")]
    pub entries: Vec<ProductEntry>,
}

#[derive(Debug, Deserialize)]
pub struct ProductDetails {
    pub machine_name: String,
    pub human_name: String,
}

impl Product {
    pub fn total_size(&self) -> u64 {
        self.entries.iter().map(|e| e.total_size()).sum()
    }
}

#[derive(Debug, Deserialize)]
pub struct ProductEntry {
    pub machine_name: String,
    pub human_name: String,

    #[serde(rename = "url")]
    pub product_details_url: String,

    pub downloads: Vec<DownloadEntry>,
}

impl ProductEntry {
    pub fn total_size(&self) -> u64 {
        self.downloads.iter().map(|e| e.total_size()).sum()
    }

    pub fn formats(&self) -> String {
        self.downloads
            .iter()
            .map(|d| d.formats())
            .collect::<Vec<_>>()
            .join(", ")
    }
}

#[derive(Debug, Deserialize)]
pub struct DownloadEntry {
    #[serde(rename = "download_struct")]
    pub sub_items: Vec<DownloadEntryItem>,
}

impl DownloadEntry {
    pub fn total_size(&self) -> u64 {
        self.sub_items.iter().map(|e| e.file_size).sum()
    }

    pub fn formats(&self) -> String {
        self.sub_items
            .iter()
            .map(|s| s.item_type.clone())
            .collect::<Vec<_>>()
            .join(", ")
    }
}

#[derive(Debug, Deserialize)]
pub struct DownloadEntryItem {
    pub md5: String,

    #[serde(rename = "name")]
    pub item_type: String,

    pub file_size: u64,

    #[serde(rename = "url")]
    pub urls: DownloadUrl,
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

    pub fn list_products(&self) -> Result<Vec<Product>, anyhow::Error> {
        let client = Client::new();

        // First: get the game keys
        let res = client
            .get("https://www.humblebundle.com/api/v1/user/order")
            .header(reqwest::header::ACCEPT, "application/json")
            .header(
                "cookie".to_owned(),
                format!("_simpleauth_sess={}", self.auth_key),
            )
            .send()?;

        if !res.status().is_success() {
            return Err(ApiError::BadHttpStatus(res.status().as_u16()).into());
        }

        let game_keys = res.json::<Vec<GameKey>>()?;

        // Second: get details for those game keys
        let query_params: Vec<_> = game_keys
            .into_iter()
            .map(|g| ("gamekeys", g.gamekey))
            .collect();

        let res = client
            .get("https://www.humblebundle.com/api/v1/orders")
            .header(reqwest::header::ACCEPT, "application/json")
            .header(
                "cookie".to_owned(),
                format!("_simpleauth_sess={}", self.auth_key),
            )
            .query(&query_params)
            .send()?;

        if !res.status().is_success() {
            return Err(ApiError::BadHttpStatus(res.status().as_u16()).into());
        }
        let product_map = res.json::<ProductMap>()?;
        Ok(product_map.into_values().collect())
    }

    pub fn read_product(&self, product_key: &str) -> Result<Product, anyhow::Error> {
        let url = format!("https://www.humblebundle.com/api/v1/order/{}", product_key);

        let client = Client::new();
        let res = client
            .get(url)
            .header(reqwest::header::ACCEPT, "application/json")
            .header(
                "cookie".to_owned(),
                format!("_simpleauth_sess={}", self.auth_key),
            )
            .send()?;

        if !res.status().is_success() {
            //eprintln!("Request failed: {}", res.status().as_u16());
            return Err(ApiError::BadHttpStatus(res.status().as_u16()).into());
        }
        res.json::<Product>().map_err(|e| e.into())
    }
}
