mod config;
pub mod download;
pub mod humble_api;
mod key_match;
pub mod util;

use anyhow::{anyhow, Context};
use clap::value_parser;
use clap::{Arg, Command};
use config::set_config;
use key_match::KeyMatch;
use std::fs;
use std::path;
use tabled::{object::Columns, Alignment, Modify, Style};

pub use config::{get_config, Config};
use humble_api::{ApiError, HumbleApi};

pub fn run() -> Result<(), anyhow::Error> {
    let list_subcommand = Command::new("list")
        .about("List all your purchased bundles")
        .arg(
        Arg::new("id-only")
            .long("id-only")
            .help("Print bundle IDs only")
            .long_help(
                "Print bundle IDs only. This can be used to chain commands together for automation.",
            ),
    ).arg(
        Arg::new("claimed")
            .long("claimed")
            .takes_value(true)
            .possible_values(["all", "yes", "no"])
            .default_value("all")
            .value_parser(value_parser!(String))
            .help("Show claimed or unclaimed bundles only. This is mostly useful if you want to know which games you have not claimed yet.")
    );

    let auth_subcommand = Command::new("auth")
        .about("Set the authentication session key")
        .long_about(
            "Set the session key used for authentication with Humble Bundle API. \
            See online documentation on how to find the session key from your web browser.",
        )
        .arg(
            Arg::new("SESSION-KEY")
                .required(true)
                .takes_value(true)
                .help("Session key that's copied from your web browser"),
        );

    let details_subcommand = Command::new("details")
        .about("Print details of a certain bundle")
        .arg(
            Arg::new("BUNDLE-KEY")
                .required(true)
                .takes_value(true)
                .help("The key for the bundle which must be shown")
                .long_help(
                    "The key for the bundle which must be shown. It can be partially entered.",
                ),
        );

    let download_subcommand = Command::new("download")
        .about("Selectively download items from a bundle")
        .arg(
            Arg::new("BUNDLE-KEY")
                .required(true)
                .help("The key for the bundle which must be downloaded")
                .long_help(
                    "The key for the bundle which must be downloaded. It can be partially entered."
                )
        )
        .arg(
            Arg::new("item-numbers")
            .short('i')
            .long("item-numbers")
            .takes_value(true)
            .help("Download only specified items")
            .long_help(
                "Download only specified items. This is a comman-separated list of item numbers to download. \
                Item numbers begin from 1 and can be a single number or a range.\n\
                Some examples:\n\n\
                '--item-numbers 1,3,5' will download items 1, 3, and 5.\n\
                '--item number 5-10' will download items 5 to 10 (inclusive)\n\n\
                When specifying ranges, either the beginning or the end of the range can be omitted.\n\
                For example, '--item-numbers 10-' will download items 10 to the end.
                "
            )
        )
        .arg(
            Arg::new("format")
                .short('f')
                .long("format")
                .takes_value(true)
                .multiple_occurrences(true)
                .help("Filter downloaded items by their format")
                .long_help(
                    "Filter downloaded files by their format. Formats are case-insensitive and \
                    this filter can be used several times to specify multiple formats.\n\n\
                    For example: --filter-by-format epub --filter-by-format mobi"
                )
        )
        .arg(
            Arg::new("max-size")
                .short('s')
                .long("max-size")
                .takes_value(true)
                .help("Filter downloaded items by their maximum size")
                .long_help(
                    "Filter downloaded items by their maximum size. This will skip any sub-item in a bundle \
                    that exceeds this limit. \
                    You can use the traditional size units such as KB or MiB. Make sure there is no space \
                    between the number and the unit. For example 14MB or 4GiB.\n\n\
                    Note: The size limit works on a sub-item level, and not per file. \
                    For example, if you specify a limit of 10 MB and a sub-item has two 6 MB books in it, \
                    this sub-items will not be downloaded, because its total size exceeds the 10 MB limit (12 MB in total)."
                    )
        );

    let sub_commands = vec![
        auth_subcommand,
        list_subcommand,
        details_subcommand,
        download_subcommand,
    ];

    let crate_name = clap::crate_name!();

    let matches = clap::Command::new(crate_name)
        .about("The missing Humble Bundle CLI")
        .version(clap::crate_version!())
        .after_help("Note: `humble-cli -h` prints a short and concise overview while `humble-cli --help` gives all details.")
        .subcommand_required(true)
        .arg_required_else_help(true)
        .subcommands(sub_commands)
        .get_matches();

    return match matches.subcommand() {
        Some(("auth", sub_matches)) => auth(sub_matches),
        Some(("details", sub_matches)) => show_bundle_details(sub_matches),
        Some(("download", sub_matches)) => download_bundle(sub_matches),
        Some(("list", sub_matches)) => list_bundles(sub_matches),
        // This shouldn't happen
        _ => Ok(()),
    };
}

fn auth(matches: &clap::ArgMatches) -> Result<(), anyhow::Error> {
    let session_key = matches.value_of("SESSION-KEY").unwrap();

    set_config(Config {
        session_key: session_key.to_owned(),
    })
}

fn handle_http_errors<T>(input: Result<T, ApiError>) -> Result<T, anyhow::Error> {
    match input {
        Ok(val) => Ok(val),
        Err(ApiError::NetworkError(e)) if e.is_status() => {
            return match e.status().unwrap() {
                reqwest::StatusCode::UNAUTHORIZED => Err(anyhow!(
                    "Unauthorized request (401). Is the session key correct?"
                )),
                reqwest::StatusCode::NOT_FOUND => Err(anyhow!(
                    "Bundle not found (404). Is the bundle key correct?"
                )),
                s => Err(anyhow!("failed with status: {}", s)),
            }
        }
        Err(e) => return Err(anyhow!("failed: {}", e)),
    }
}

fn list_bundles(matches: &clap::ArgMatches) -> Result<(), anyhow::Error> {
    let id_only = matches.is_present("id-only");
    // It has a default value, so calling unwrap is safw
    let claimed_filter = matches.get_one::<String>("claimed").unwrap();

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
        bundles = bundles
            .into_iter()
            .filter(|b| b.is_fully_claimed() == claimed)
            .collect();
    }

    if id_only {
        for b in bundles {
            println!("{}", b.gamekey);
        }

        return Ok(());
    }

    println!("{} bundle(s) found.", bundles.len());

    if bundles.len() == 0 {
        return Ok(());
    }

    let mut builder = tabled::builder::Builder::default().set_columns([
        "Key",
        "Name",
        "Size",
        "Claimed",
    ]);
    for p in bundles {
        builder = builder.add_record([
            p.gamekey.as_str(),
            p.details.human_name.as_str(),
            util::humanize_bytes(p.total_size()).as_str(),
            if p.is_fully_claimed() { "Yes" } else { "No" },
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

fn show_bundle_details(matches: &clap::ArgMatches) -> Result<(), anyhow::Error> {
    let config = get_config()?;
    let bundle_key = matches.value_of("BUNDLE-KEY").unwrap();
    let api = crate::HumbleApi::new(&config.session_key);

    let bundle_key = match find_key(handle_http_errors(api.list_bundle_keys())?, bundle_key) {
        Some(key) => key,
        None => return Ok(()),
    };

    let bundle = handle_http_errors(api.read_bundle(&bundle_key))?;

    println!();
    println!("{}", bundle.details.human_name);
    println!("Purchased: {}", bundle.created.format("%v %I:%M %p"));
    println!("Total size: {}", util::humanize_bytes(bundle.total_size()));
    println!();
    if bundle.has_unused_tpks() {
        println!("This bundle has keys that can be redeemed!");
    }

    println!();

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

    Ok(())
}

fn download_bundle(matches: &clap::ArgMatches) -> Result<(), anyhow::Error> {
    let config = get_config()?;
    let bundle_key = matches.value_of("BUNDLE-KEY").unwrap();
    let formats = if let Some(values) = matches.values_of("format") {
        values.map(|f| f.to_lowercase()).collect::<Vec<_>>()
    } else {
        vec![]
    };

    let max_size: u64 = if let Some(byte_str) = matches.value_of("max-size") {
        util::byte_string_to_number(byte_str)
            .context(format!("failed to parse the specified size: {}", byte_str))?
    } else {
        0
    };

    let api = crate::HumbleApi::new(&config.session_key);

    let bundle_key = match find_key(handle_http_errors(api.list_bundle_keys())?, bundle_key) {
        Some(key) => key,
        None => return Ok(()),
    };

    let bundle = handle_http_errors(api.read_bundle(&bundle_key))?;

    // To parse the item number ranges, we need to know the max value
    // for unbounded ranges (e.g. 12-). That's why we parse this argument
    // after we read the bundle from the API.
    let item_numbers = if let Some(value) = matches.value_of("item-numbers") {
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
