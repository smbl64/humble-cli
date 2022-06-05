#![allow(dead_code)]

use clap::{arg, command, Command};
use tabled::{object::Columns, Alignment, Modify, Style};

fn main() -> Result<(), anyhow::Error> {
    let session_key = std::fs::read_to_string("session.key")
        .expect("failed to read the session key from `session.key` file");

    let session_key = session_key.trim_end();

    let list_subcommand = Command::new("list")
        .about("List purchases")
        .arg(arg!([NAME]));

    let download_subcommand = Command::new("download")
        .about("Download all items in a product")
        .arg(arg!([KEY]));

    let sub_commands = vec![list_subcommand, download_subcommand];

    let matches = command!()
        .propagate_version(true)
        .subcommand_required(true)
        .arg_required_else_help(true)
        .subcommands(sub_commands)
        .get_matches();

    match matches.subcommand() {
        Some(("list", _sub_matches)) => list_products(&session_key),
        _ => {}
    }

    Ok(())
}

fn list_products(session_key: &str) {
    let api = humble_cli::HumbleApi::new(session_key);
    let products = humble_cli::run_future(api.list_products()).unwrap();
    println!("Done: {} products", products.len());
    let mut builder = tabled::builder::Builder::default().set_columns(["Key", "Name", "Size"]);
    for p in products {
        builder = builder.add_record([
            &p.gamekey,
            &p.details.human_name,
            &bytesize::to_string(p.total_size(), true),
        ]);
    }

    let table = builder
        .build()
        .with(Style::psql())
        .with(Modify::new(Columns::single(1)).with(Alignment::left()))
        .with(Modify::new(Columns::single(2)).with(Alignment::right()));
    println!("{table}");
}
