pub mod download;
pub mod humble_api;
pub mod util;

use anyhow::{anyhow, Context};
use clap::{Arg, Command};
use std::fs;
use std::path;
use tabled::{object::Columns, Alignment, Modify, Style};

use humble_api::{ApiError, HumbleApi};

#[derive(Debug)]
pub struct Config {
    pub session_key: String,
}

pub fn get_config() -> Result<Config, anyhow::Error> {
    let session_key = std::fs::read_to_string("session.key")
        .context("failed to read the session key from `session.key` file")?;

    let session_key = session_key.trim_end().to_owned();

    Ok(Config { session_key })
}

pub fn run(config: Config) -> Result<(), anyhow::Error> {
    let list_subcommand = Command::new("list").about("List all purchased bundles");

    let details_subcommand = Command::new("details")
        .about("Print details of a certain bundle")
        .arg(
            Arg::new("KEY")
                .required(true)
                .takes_value(true)
                .help("Bundle's key"),
        );

    let download_subcommand = Command::new("download")
        .about("Download all items in a bundle")
        .arg(
            Arg::new("KEY")
                .required(true)
                .help("The bundle which must be downloaded"),
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

    let sub_commands = vec![list_subcommand, details_subcommand, download_subcommand];

    let crate_name = clap::crate_name!();

    let matches = clap::Command::new(crate_name)
        .about("The missing Humble Bundle CLI")
        .version(clap::crate_version!())
        .after_help("Note: `humble-cli -h` prints a short and concise overview while `humble-cli --help` gives all details.")
        .propagate_version(true)
        .subcommand_required(true)
        .arg_required_else_help(true)
        .subcommands(sub_commands)
        .get_matches();

    return match matches.subcommand() {
        Some(("list", _)) => list_bundles(config),
        Some(("details", sub_matches)) => show_bundle_details(config, sub_matches),
        Some(("download", sub_matches)) => download_bundle(config, sub_matches),
        // This shouldn't happen
        _ => Ok(()),
    };
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

fn list_bundles(config: Config) -> Result<(), anyhow::Error> {
    let api = HumbleApi::new(&config.session_key);
    let bundles = handle_http_errors(api.list_bundles())?;

    println!("{} bundles(s) found.", bundles.len());

    let mut builder = tabled::builder::Builder::default().set_columns(["Key", "Name", "Size"]);
    for p in bundles {
        builder = builder.add_record([
            &p.gamekey,
            &p.details.human_name,
            &util::humanize_bytes(p.total_size()),
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

fn show_bundle_details(config: Config, matches: &clap::ArgMatches) -> Result<(), anyhow::Error> {
    let bundle_key = matches.value_of("KEY").unwrap();
    let api = crate::HumbleApi::new(&config.session_key);
    let bundle = handle_http_errors(api.read_bundle(bundle_key))?;

    println!("");
    println!("{}", bundle.details.human_name);
    println!("Total size: {}", util::humanize_bytes(bundle.total_size()));
    println!("");

    let mut builder = tabled::builder::Builder::default();
    builder = builder.set_columns(["", "Sub-item", "Format", "Total Size"]);

    for (idx, entry) in bundle.entries.iter().enumerate() {
        builder = builder.add_record([
            &idx.to_string(),
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

fn download_bundle(config: Config, matches: &clap::ArgMatches) -> Result<(), anyhow::Error> {
    let bundle_key = matches.value_of("KEY").unwrap();
    let formats = if let Some(values) = matches.values_of("format") {
        values.collect::<Vec<_>>()
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
    let bundle = handle_http_errors(api.read_bundle(bundle_key))?;

    let dir_name = util::replace_invalid_chars_in_filename(&bundle.details.human_name);
    let bundle_dir = create_dir(&dir_name)?;

    let client = reqwest::Client::new();

    for product in bundle.entries {
        if max_size > 0 && product.total_size() > max_size {
            continue;
        }

        //product_entry.formats_as_vec()

        println!("");
        println!("{}", product.human_name);

        let entry_dir = bundle_dir.join(product.human_name);
        if !entry_dir.exists() {
            fs::create_dir(&entry_dir)?;
        }

        for download_entry in product.downloads {
            for ele in download_entry.sub_items {
                let filename = util::extract_filename_from_url(&ele.url.web)
                    .context(format!("Cannot get file name from URL '{}'", &ele.url.web))?;
                let download_path = entry_dir.join(&filename);

                let f = download::download_file(
                    &client,
                    &ele.url.web,
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
