use humble_cli::run;

fn main() {
    let crate_name = env!("CARGO_PKG_NAME");
    if let Err(e) = run() {
        eprintln!("{}: {:?}", crate_name, e);
        std::process::exit(1);
    }
}
