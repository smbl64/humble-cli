use humble_cli::{get_config, run};

fn main() {
    let crate_name = env!("CARGO_PKG_NAME");
    if let Err(e) = get_config().and_then(run) {
        eprintln!("{}: {:?}", crate_name, e);
        std::process::exit(1);
    }
}
