use std::io;

use anyhow::Context;
use clap::{builder::ValueParser, value_parser, Arg, Command};
use clap_complete::Shell;
use humble_cli::{download_bundles, prelude::*};

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

fn parse_match_mode(input: &str) -> Result<MatchMode, anyhow::Error> {
    MatchMode::try_from(input).map_err(|e| anyhow::anyhow!(e))
}

fn run() -> Result<(), anyhow::Error> {
    let list_subcommand = Command::new("list")
        .about("List all your purchased bundles")
        .visible_alias("ls")
        .arg(
        Arg::new("field")
            .long("field")
            .takes_value(true)
            .multiple_occurrences(true)
            .help("Print bundle with the specified fields only")
            .long_help(
                "Print bundle with the specified fields only. This can be used to chain commands together for automation. \
                 If fields are not set, all fields will be printed  \
                 Valid Fields: [key, name, size, claimed] \
                 Use example: --field key --field name ",
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

    let completion_subcommand = Command::new("completion")
        .about("Generate shell completions")
        .arg(
            Arg::new("SHELL")
                .help("Shell type to generate completions for")
                .possible_values(["bash", "elvish", "fish", "powershell", "zsh"])
                .takes_value(true)
                .required(true)
                .value_parser(value_parser!(Shell)),
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
        .visible_alias("info")
        .arg(
            Arg::new("BUNDLE-KEY")
                .required(true)
                .takes_value(true)
                .help("The key for the bundle which must be shown")
                .long_help(
                    "The key for the bundle which must be shown. It can be partially entered.",
                ),
        );

    let search_subcommand = Command::new("search")
        .about("Search through all bundle products for keywords")
        .arg(
            Arg::new("KEYWORDS")
                .required(true)
                .action(clap::ArgAction::Append)
                .multiple_values(true)
                .help("Search keywords"),
        )
        .arg(
            Arg::new("mode")
                .long("mode")
                .value_name("mode")
                .takes_value(true)
                .possible_values(["all", "any"])
                .default_value("any")
                .value_parser(ValueParser::new(parse_match_mode))
                .help("Whether all or any of the keywords should match the name"),
        );
    let download_subcommand = Command::new("download")
        .about("Selectively download items from a bundle")
        .visible_alias("d")
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
            Arg::new("torrents")
                .short('t')
                .long("torrents")
                .takes_value(false)
                .help("Download only .torrent files for items")
                .long_help(
                    "Download only the BitTorrent files for the given items. This will prevent \
                    all the original files from downloading. To download both, run again without \
                    this flag."
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
        ).arg(
            Arg::new("cur-dir")
                .short('c')
                .long("cur-dir")
                .takes_value(false)
                .help("Download into current dir")
                .long_help(
                    "One Directoy for each entry is created, \
                    but no bundle directory is created."
                )
        );

    let bulk_download_subcommand = Command::new("bulk-download")
        .about("Selectively download items from a bundle")
        .visible_alias("b")
        .arg(
            Arg::new("INPUT-FILE")
                .required(true)
                .help("Takes a list input file")
                .long_help(
                    "This takes the input created from the list command, then iterates all items, \
                    Using the bundle name as directory name")
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
            Arg::new("torrents")
                .short('t')
                .long("torrents")
                .takes_value(false)
                .help("Download only .torrent files for items")
                .long_help(
                    "Download only the BitTorrent files for the given items. This will prevent \
                    all the original files from downloading. To download both, run again without \
                    this flag."
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
        ).arg(
            Arg::new("cur-dir")
                .short('c')
                .long("cur-dir")
                .takes_value(false)
                .help("Download into current dir")
                .long_help(
                    "One Directoy for each entry is created, \
                    but no bundle directory is created."
                )
        );

    let sub_commands = vec![
        auth_subcommand,
        list_subcommand,
        list_choices_subcommand,
        details_subcommand,
        download_subcommand,
        search_subcommand,
        completion_subcommand,
        bulk_download_subcommand,
    ];

    let crate_name = clap::crate_name!();

    let mut root = clap::Command::new(crate_name)
        .about("The missing Humble Bundle CLI")
        .version(clap::crate_version!())
        .after_help("Note: `humble-cli -h` prints a short and concise overview while `humble-cli --help` gives all details.")
        .subcommand_required(true)
        .arg_required_else_help(true)
        .subcommands(sub_commands);

    let matches = root.clone().get_matches();
    match matches.subcommand() {
        Some(("completion", sub_matches)) => {
            if let Some(g) = sub_matches.get_one::<Shell>("SHELL").copied() {
                let crate_name = clap::crate_name!();
                clap_complete::generate(g, &mut root, crate_name, &mut io::stdout());
            }
            Ok(())
        }
        Some(("auth", sub_matches)) => {
            let session_key = sub_matches.value_of("SESSION-KEY").unwrap();
            auth(session_key)
        }
        Some(("details", sub_matches)) => {
            let bundle_key = sub_matches.value_of("BUNDLE-KEY").unwrap();
            show_bundle_details(bundle_key)
        }
        Some(("search", sub_matches)) => {
            let keywords: Vec<String> =
                sub_matches.get_many("KEYWORDS").unwrap().cloned().collect();
            let keywords = keywords.join(" ");

            let match_mode: &MatchMode = sub_matches.get_one("mode").unwrap();
            // let keywords = sub_matches.value_of("KEYWORDS").unwrap();
            search(&keywords, *match_mode)
        }
        Some(("download", sub_matches)) => {
            let bundle_key = sub_matches.value_of("BUNDLE-KEY").unwrap();
            let formats = if let Some(values) = sub_matches.values_of("format") {
                values.map(|f| f.to_lowercase()).collect::<Vec<_>>()
            } else {
                vec![]
            };
            let max_size: u64 = if let Some(byte_str) = sub_matches.value_of("max-size") {
                byte_string_to_number(byte_str)
                    .context(format!("failed to parse the specified size: {}", byte_str))?
            } else {
                0
            };
            let item_numbers = sub_matches.value_of("item-numbers");
            let torrents_only = sub_matches.is_present("torrents");
            let cur_dir = sub_matches.is_present("cur-dir");
            download_bundle(
                bundle_key,
                &formats,
                max_size,
                item_numbers,
                torrents_only,
                cur_dir,
            )
        }
        Some(("list", sub_matches)) => {
            let fields = if let Some(values) = sub_matches.values_of("field") {
                values.map(|f| f.to_lowercase()).collect::<Vec<_>>()
            } else {
                vec![]
            };
            let claimed_filter = sub_matches
                .get_one::<String>("claimed")
                .map(String::as_str)
                .unwrap_or("all");
            list_bundles(fields, claimed_filter)
        }
        Some(("list-choices", sub_matches)) => {
            let period: &ChoicePeriod = sub_matches.get_one("period").unwrap();
            list_humble_choices(period)
        }
        Some(("bulk-download", sub_matches)) => {
            let bundle_file = sub_matches.value_of("INPUT-FILE").unwrap();
            let formats = if let Some(values) = sub_matches.values_of("format") {
                values.map(|f| f.to_lowercase()).collect::<Vec<_>>()
            } else {
                vec![]
            };
            let max_size: u64 = if let Some(byte_str) = sub_matches.value_of("max-size") {
                byte_string_to_number(byte_str)
                    .context(format!("failed to parse the specified size: {}", byte_str))?
            } else {
                0
            };
            let torrents_only = sub_matches.is_present("torrents");
            let cur_dir = sub_matches.is_present("cur-dir");
            download_bundles(bundle_file, formats, max_size, torrents_only, cur_dir)
        }
        // This shouldn't happen
        _ => Ok(()),
    }
}
