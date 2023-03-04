use std::collections::{BTreeMap, HashMap};

use chrono::NaiveDateTime;
use serde::Deserialize;
use serde_with::{serde_as, VecSkipError};

#[derive(Debug, PartialEq)]
pub enum ClaimStatus {
    Yes,
    No,
    NotAvailable,
}

impl ToString for ClaimStatus {
    fn to_string(&self) -> String {
        match self {
            Self::Yes => "Yes",
            Self::No => "No",
            Self::NotAvailable => "-",
        }
        .to_owned()
    }
}

// ===========================================================================
// Models related to the purchased Bundles
// ===========================================================================
pub type BundleMap = HashMap<String, Bundle>;

#[serde_as]
#[derive(Debug, Deserialize)]
pub struct Bundle {
    pub gamekey: String,
    pub created: NaiveDateTime,
    pub claimed: bool,

    pub tpkd_dict: HashMap<String, serde_json::Value>,

    #[serde(rename = "product")]
    pub details: BundleDetails,

    #[serde(rename = "subproducts")]
    #[serde_as(as = "VecSkipError<_>")]
    pub products: Vec<Product>,
}

pub struct ProductKey {
    pub redeemed: bool,
    pub human_name: String,
}

impl Bundle {
    pub fn claim_status(&self) -> ClaimStatus {
        let product_keys = self.product_keys();
        let total_count = product_keys.len();
        if total_count == 0 {
            return ClaimStatus::NotAvailable;
        }

        let unused_count = product_keys.iter().filter(|k| !k.redeemed).count();
        if unused_count > 0 {
            ClaimStatus::No
        } else {
            ClaimStatus::Yes
        }
    }

    pub fn product_keys(&self) -> Vec<ProductKey> {
        let Some(tpks) = self.tpkd_dict.get("all_tpks") else {
            return vec![];
        };

        let tpks = tpks.as_array().expect("cannot read all_tpks");

        let mut result = vec![];
        for tpk in tpks {
            let redeemed = tpk["redeemed_key_val"].is_string();
            let human_name = tpk["human_name"]
                .as_str()
                .expect("expected human_name to be a string")
                .to_owned();

            result.push(ProductKey {
                redeemed,
                human_name,
            });
        }

        result
    }
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
pub struct GameKey {
    pub gamekey: String,
}

// ===========================================================================
// Models related to the Bundle Choices
// ===========================================================================
#[derive(Debug, Deserialize)]
pub struct HumbleChoice {
    #[serde(rename = "contentChoiceOptions")]
    pub options: ContentChoiceOptions,
}

#[derive(Debug, Deserialize)]
pub struct ContentChoiceOptions {
    #[serde(rename = "contentChoiceData")]
    pub data: ContentChoiceData,

    pub gamekey: Option<String>,

    #[serde(rename = "isActiveContent")]
    pub is_active_content: bool,

    pub title: String,
}

#[derive(Debug, Deserialize)]
pub struct ContentChoiceData {
    pub game_data: BTreeMap<String, GameData>,
}

#[derive(Debug, Deserialize)]
pub struct GameData {
    pub title: String,
    pub tpkds: Vec<Tpkd>,
}

#[derive(Debug, Deserialize)]
pub struct Tpkd {
    pub gamekey: Option<String>,
    pub human_name: String,
    pub redeemed_key_val: Option<String>,
}

impl Tpkd {
    pub fn claim_status(&self) -> ClaimStatus {
        let redeemed = self.redeemed_key_val.is_some();
        let is_active = self.gamekey.is_some();
        if is_active && redeemed {
            ClaimStatus::Yes
        } else if is_active {
            ClaimStatus::No
        } else {
            ClaimStatus::NotAvailable
        }
    }
}
