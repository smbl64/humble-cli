use futures_util::StreamExt;
use indicatif::{ProgressBar, ProgressStyle};
use reqwest::Client;
use std::cmp::min;
use std::fs::File;
use std::io::{Seek, Write};

#[derive(Debug, thiserror::Error)]
pub enum DownloadError {
    #[error(transparent)]
    NetworkError(#[from] reqwest::Error),

    #[error(transparent)]
    IOError(#[from] std::io::Error),

    #[error("{0}")]
    GenericError(String),
}

impl DownloadError {
    fn from_string(s: String) -> Self {
        DownloadError::GenericError(s)
    }
}

pub async fn download_file(
    client: &Client,
    url: &str,
    path: &str,
    title: &str,
) -> Result<(), DownloadError> {
    let res = client.get(url).send().await?;

    let total_size = res
        .content_length()
        .ok_or(DownloadError::from_string(format!(
            "Failed to get content length from '{}'",
            &url
        )))?;

    let mut file;
    let mut downloaded: u64 = 0;
    let mut stream = res.bytes_stream();

    if std::path::Path::new(path).exists() {
        file = std::fs::OpenOptions::new()
            .read(true)
            .append(true)
            .open(path)
            .unwrap();

        let file_size = std::fs::metadata(path).unwrap().len();
        file.seek(std::io::SeekFrom::Start(file_size))?;
        downloaded = file_size;

        if downloaded >= total_size {
            println!("  Nothing to do. File already exists.");
            return Ok(());
        }
    } else {
        file = File::create(path)?;
    }

    let pb = get_progress_bar(total_size);
    pb.set_message(format!("Downloading {}", title));

    while let Some(item) = stream.next().await {
        let chunk = item?;
        file.write(&chunk)?;

        let new = min(downloaded + (chunk.len() as u64), total_size);
        downloaded = new;
        pb.set_position(new);
    }

    pb.finish_and_clear();
    println!("  Downloaded {}", title);
    Ok(())
}

fn get_progress_bar(total_size: u64) -> ProgressBar {
    let pb = ProgressBar::new(total_size);
    let pb_template =
        "  {msg}\n  {spinner:.green} [{elapsed}] [{bar}] {bytes} / {total_bytes} ({bytes_per_sec})";

    pb.set_style(
        ProgressStyle::default_bar()
            .template(pb_template)
            .progress_chars("=> "),
    );
    pb
}
