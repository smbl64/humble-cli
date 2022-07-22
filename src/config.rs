use std::path::PathBuf;

use anyhow::{anyhow, Context};

#[derive(Debug)]
pub struct Config {
    pub session_key: String,
}

pub fn get_config() -> Result<Config, anyhow::Error> {
    let file_name = get_config_file_name()?;
    let session_key = std::fs::read_to_string(&file_name).with_context(|| {
        format!(
            "failed to read the session key from `{}` file",
            &file_name.to_str().unwrap()
        )
    })?;

    let session_key = session_key.trim_end().to_owned();

    Ok(Config { session_key })
}

pub fn set_config(config: Config) -> Result<(), anyhow::Error> {
    let file_name = get_config_file_name()?;

    std::fs::write(file_name, config.session_key)?;

    Ok(())
}

fn get_config_file_name() -> anyhow::Result<PathBuf> {
    let mut home = dirs::home_dir().ok_or_else(|| anyhow!("cannot find the home directory"))?;
    home.push(".humble-cli-key");
    Ok(home)
}
