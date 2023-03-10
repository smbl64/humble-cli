mod config;
mod download;
mod humble_api;
mod key_match;
mod models;
mod util;

pub mod prelude {
    pub use crate::auth;
    pub use crate::download_bundle;
    pub use crate::list_bundles;
    pub use crate::list_humble_choices;
    pub use crate::show_bundle_details;

    pub use crate::humble_api::{ApiError, HumbleApi};
    pub use crate::models::*;
    pub use crate::util::byte_string_to_number;
}

use crate::models::ClaimStatus;
use anyhow::{anyhow, Context};
use config::{get_config, set_config, Config};
use humble_api::{ApiError, HumbleApi};
use key_match::KeyMatch;
use std::fs;
use std::path;
use tabled::{object::Columns, Alignment, Modify, Style};

pub fn auth(session_key: &str) -> Result<(), anyhow::Error> {
    set_config(Config {
        session_key: session_key.to_owned(),
    })
}

pub fn handle_http_errors<T>(input: Result<T, ApiError>) -> Result<T, anyhow::Error> {
    match input {
        Ok(val) => Ok(val),
        Err(ApiError::NetworkError(e)) if e.is_status() => match e.status().unwrap() {
            reqwest::StatusCode::UNAUTHORIZED => Err(anyhow!(
                "Unauthorized request (401). Is the session key correct?"
            )),
            reqwest::StatusCode::NOT_FOUND => Err(anyhow!(
                "Bundle not found (404). Is the bundle key correct?"
            )),
            s => Err(anyhow!("failed with status: {}", s)),
        },
        Err(e) => Err(anyhow!("failed: {}", e)),
    }
}

pub fn list_humble_choices(period: &str) -> Result<(), anyhow::Error> {
    let config = get_config()?;
    let api = HumbleApi::new(&config.session_key);

    let choices = api.read_bundle_choices(period)?;

    println!();
    println!("{}", choices.options.title);
    println!();

    let options = choices.options;

    let mut builder = tabled::builder::Builder::default().set_columns(["#", "Title", "Redeemed"]);

    let mut counter = 1;
    let mut all_redeemed = true;
    for (_, game_data) in options.data.game_data.iter() {
        for tpkd in game_data.tpkds.iter() {
            builder = builder.add_record([
                counter.to_string().as_str(),
                tpkd.human_name.as_str(),
                tpkd.claim_status().to_string().as_str(),
            ]);

            counter += 1;

            if tpkd.claim_status() == ClaimStatus::No {
                all_redeemed = false;
            }
        }
    }

    let table = builder
        .build()
        .with(Style::psql())
        .with(Modify::new(Columns::single(0)).with(Alignment::right()))
        .with(Modify::new(Columns::single(1)).with(Alignment::left()));

    println!("{table}");

    if !all_redeemed {
        let url = "https://www.humblebundle.com/membership/home";
        println!("Visit {url} to redeem your keys.");
    }
    Ok(())
}

pub fn list_bundles(id_only: bool, claimed_filter: &str) -> Result<(), anyhow::Error> {
    let config = get_config()?;
    let api = HumbleApi::new(&config.session_key);

    // If no filter is required, we can do a single call
    // and finish quickly. Otherwise we will need to fetch
    // all bundle data and filter them.
    if id_only && claimed_filter == "all" {
        let ids = handle_http_errors(api.list_bundle_keys())?;
        for id in ids {
            println!("{}", id);
        }

        return Ok(());
    }

    let mut bundles = handle_http_errors(api.list_bundles())?;

    if claimed_filter != "all" {
        let claimed = claimed_filter == "yes";
        bundles.retain(|b| {
            let status = b.claim_status();
            status == ClaimStatus::Yes && claimed || status == ClaimStatus::No && !claimed
        });
    }

    if id_only {
        for b in bundles {
            println!("{}", b.gamekey);
        }

        return Ok(());
    }

    println!("{} bundle(s) found.", bundles.len());

    if bundles.is_empty() {
        return Ok(());
    }

    let mut builder =
        tabled::builder::Builder::default().set_columns(["Key", "Name", "Size", "Claimed"]);
    for p in bundles {
        builder = builder.add_record([
            p.gamekey.as_str(),
            p.details.human_name.as_str(),
            util::humanize_bytes(p.total_size()).as_str(),
            p.claim_status().to_string().as_str(),
        ]);
    }

    let table = builder
        .build()
        .with(Style::psql())
        .with(Modify::new(Columns::single(1)).with(Alignment::left()))
        .with(Modify::new(Columns::single(2)).with(Alignment::right()));
    println!("{table}");

    Ok(())
}

fn find_key(all_keys: Vec<String>, key_to_find: &str) -> Option<String> {
    let key_match = KeyMatch::new(all_keys, key_to_find);
    let keys = key_match.get_matches();

    match keys.len() {
        1 => Some(keys[0].clone()),
        0 => {
            eprintln!("No bundle matches '{}'", key_to_find);
            None
        }
        _ => {
            eprintln!("More than one bundle matches '{}':", key_to_find);
            for key in keys {
                eprintln!("{}", key);
            }
            None
        }
    }
}

pub fn show_bundle_details(bundle_key: &str) -> Result<(), anyhow::Error> {
    let config = get_config()?;
    let api = crate::HumbleApi::new(&config.session_key);

    let bundle_key = match find_key(handle_http_errors(api.list_bundle_keys())?, bundle_key) {
        Some(key) => key,
        None => return Ok(()),
    };

    let bundle = handle_http_errors(api.read_bundle(&bundle_key))?;

    println!();
    println!("{}", bundle.details.human_name);
    println!();
    println!("Purchased  : {}", bundle.created.format("%v %I:%M %p"));
    println!("Total size : {}", util::humanize_bytes(bundle.total_size()));
    println!();

    if !bundle.products.is_empty() {
        let mut builder = tabled::builder::Builder::default();
        builder = builder.set_columns(["#", "Sub-item", "Format", "Total Size"]);

        for (idx, entry) in bundle.products.iter().enumerate() {
            builder = builder.add_record([
                &(idx + 1).to_string(),
                &entry.human_name,
                &entry.formats(),
                &util::humanize_bytes(entry.total_size()),
            ]);
        }
        let table = builder
            .build()
            .with(Style::psql())
            .with(Modify::new(Columns::single(0)).with(Alignment::right()))
            .with(Modify::new(Columns::single(1)).with(Alignment::left()))
            .with(Modify::new(Columns::single(2)).with(Alignment::left()))
            .with(Modify::new(Columns::single(3)).with(Alignment::right()));
        println!("{table}");
    } else {
        println!("No items to show.");
    }

    // Product keys
    let product_keys = bundle.product_keys();
    if !product_keys.is_empty() {
        println!();
        println!("Keys in this bundle:");
        println!();
        let mut builder = tabled::builder::Builder::default();
        builder = builder.set_columns(["#", "Key Name", "Redeemed"]);

        let mut all_redeemed = true;
        for (idx, entry) in product_keys.iter().enumerate() {
            builder = builder.add_record([
                (idx + 1).to_string().as_str(),
                entry.human_name.as_str(),
                if entry.redeemed { "Yes" } else { "No" },
            ]);

            if !entry.redeemed {
                all_redeemed = false;
            }
        }

        let table = builder
            .build()
            .with(Style::psql())
            .with(Modify::new(Columns::single(0)).with(Alignment::right()))
            .with(Modify::new(Columns::single(1)).with(Alignment::left()))
            .with(Modify::new(Columns::single(2)).with(Alignment::center()));

        println!("{table}");

        if !all_redeemed {
            let url = "https://www.humblebundle.com/home/keys";
            println!("Visit {url} to redeem your keys.");
        }
    }

    Ok(())
}

pub fn download_bundle(
    bundle_key: &str,
    formats: Vec<String>,
    max_size: u64,
    item_numbers: Option<&str>,
) -> Result<(), anyhow::Error> {
    let config = get_config()?;

    let api = crate::HumbleApi::new(&config.session_key);

    let bundle_key = match find_key(handle_http_errors(api.list_bundle_keys())?, bundle_key) {
        Some(key) => key,
        None => return Ok(()),
    };

    let bundle = handle_http_errors(api.read_bundle(&bundle_key))?;

    // To parse the item number ranges, we need to know the max value
    // for unbounded ranges (e.g. 12-). That's why we parse this argument
    // after we read the bundle from the API.
    let item_numbers = if let Some(value) = item_numbers {
        let ranges = value.split(',').collect::<Vec<_>>();
        util::union_usize_ranges(&ranges, bundle.products.len())?
    } else {
        vec![]
    };

    // Filter products based on entered criteria
    // Note that item numbers entered by user start at 1, while our index
    // starts as 0.
    let products = bundle
        .products
        .iter()
        .enumerate()
        .filter(|&(i, _)| item_numbers.is_empty() || item_numbers.contains(&(i + 1)))
        .map(|(_, p)| p)
        .filter(|p| max_size == 0 || p.total_size() < max_size)
        .filter(|p| {
            formats.is_empty() || util::str_vectors_intersect(&p.formats_as_vec(), &formats)
        })
        .collect::<Vec<_>>();

    if products.is_empty() {
        println!("Nothing to download");
        return Ok(());
    }

    // Create the bundle directory
    let dir_name = util::replace_invalid_chars_in_filename(&bundle.details.human_name);
    let bundle_dir = create_dir(&dir_name)?;

    let client = reqwest::Client::new();

    for product in products {
        if max_size > 0 && product.total_size() > max_size {
            continue;
        }

        println!();
        println!("{}", product.human_name);

        let dir_name = util::replace_invalid_chars_in_filename(&product.human_name);
        let entry_dir = bundle_dir.join(dir_name);
        if !entry_dir.exists() {
            fs::create_dir(&entry_dir)?;
        }

        for product_download in product.downloads.iter() {
            for dl_info in product_download.items.iter() {
                if !formats.is_empty() && !formats.contains(&dl_info.format.to_lowercase()) {
                    println!("Skipping '{}'", dl_info.format);
                    continue;
                }

                let filename = util::extract_filename_from_url(&dl_info.url.web).context(
                    format!("Cannot get file name from URL '{}'", &dl_info.url.web),
                )?;
                let download_path = entry_dir.join(&filename);

                let f = download::download_file(
                    &client,
                    &dl_info.url.web,
                    download_path.to_str().unwrap(),
                    &filename,
                );
                util::run_future(f)?;
            }
        }
    }

    Ok(())
}

fn create_dir(dir: &str) -> Result<path::PathBuf, std::io::Error> {
    let dir = path::Path::new(dir).to_owned();
    if !dir.exists() {
        fs::create_dir(&dir)?;
    }
    Ok(dir)
}
