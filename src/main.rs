use anyhow::Context;
use clap::{builder::ValueParser, value_parser, Arg, Command};
use humble_cli::prelude::*;

fn main() {
    let crate_name = env!("CARGO_PKG_NAME");
    if let Err(e) = run() {
        eprintln!("{}: {:?}", crate_name, e);
        std::process::exit(1);
    }
}

fn parse_choices_period(input: &str) -> Result<ChoicePeriod, anyhow::Error> {
    ChoicePeriod::try_from(input).map_err(|e| anyhow::anyhow!(e))
}

fn run() -> Result<(), anyhow::Error> {
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
            .value_name("value")
            .takes_value(true)
            .possible_values(["all", "yes", "no"])
            .default_value("all")
            .value_parser(value_parser!(String))
            .help("Show claimed or unclaimed bundles only")
            .long_help(
                "Show claimed or unclaimed bundles only. \
                    This is useful if you want to know which games or bundles you have not claimed yet."
            )
    );

    let list_choices_subcommand = Command::new("list-choices")
        .about("List your current Humble Choices")
        .arg(
            Arg::new("period")
                .default_value("current")
                .value_parser(ValueParser::new(parse_choices_period))
                .help("The month and the year to use for search. For example: 'january-2023'.\nUse 'current' for the current month."),
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
        list_choices_subcommand,
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
        Some(("auth", sub_matches)) => {
            let session_key = sub_matches.value_of("SESSION-KEY").unwrap();
            auth(session_key)
        }
        Some(("details", sub_matches)) => {
            let bundle_key = sub_matches.value_of("BUNDLE-KEY").unwrap();
            show_bundle_details(bundle_key)
        }
        Some(("download", sub_matches)) => {
            let bundle_key = sub_matches.value_of("BUNDLE-KEY").unwrap();
            let formats = if let Some(values) = matches.values_of("format") {
                values.map(|f| f.to_lowercase()).collect::<Vec<_>>()
            } else {
                vec![]
            };
            let max_size: u64 = if let Some(byte_str) = matches.value_of("max-size") {
                byte_string_to_number(byte_str)
                    .context(format!("failed to parse the specified size: {}", byte_str))?
            } else {
                0
            };
            let item_numbers = matches.value_of("item-numbers");
            download_bundle(bundle_key, formats, max_size, item_numbers)
        }
        Some(("list", sub_matches)) => {
            let id_only = sub_matches.is_present("id-only");
            let claimed_filter = sub_matches
                .get_one::<String>("claimed")
                .map(String::as_str)
                .unwrap_or("all");
            list_bundles(id_only, claimed_filter)
        }
        Some(("list-choices", sub_matches)) => {
            let period: &ChoicePeriod = sub_matches.get_one("period").unwrap();
            list_humble_choices(period)
        }

        // This shouldn't happen
        _ => Ok(()),
    };
}
