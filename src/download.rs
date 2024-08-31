use futures_util::StreamExt;
use indicatif::{ProgressBar, ProgressStyle};
use reqwest::Client;
use std::cmp::min;
use std::fs::File;
use std::io::{Seek, Write};
use std::time::Duration;

#[derive(Debug, thiserror::Error)]
pub enum DownloadError {
    #[error(transparent)]
    Network(#[from] reqwest::Error),

    #[error(transparent)]
    IO(#[from] std::io::Error),

    #[error("{0}")]
    Generic(String),
}

impl DownloadError {
    fn from_string(s: String) -> Self {
        DownloadError::Generic(s)
    }
}

pub async fn download_file(
    client: &Client,
    url: &str,
    path: &str,
    title: &str,
) -> Result<(), DownloadError> {
    const RETRY_SECONDS: u64 = 5;
    let mut retries = 3;

    loop {
        let res = _download_file(client, url, path, title).await;

        retries -= 1;
        if retries < 0 {
            return res;
        }

        match res {
            Err(DownloadError::Network(ref net_err))
                if net_err.is_connect() || net_err.is_timeout() =>
            {
                println!("  Will retry in {} seconds...", RETRY_SECONDS);
                tokio::time::sleep(Duration::from_secs(RETRY_SECONDS)).await;
                continue;
            }
            _ => return res,
        };
    }
}

async fn _download_file(
    client: &Client,
    url: &str,
    path: &str,
    title: &str,
) -> Result<(), DownloadError> {
    let (mut file, mut downloaded) = open_file_for_write(path)?;
    let total_size = get_content_length(client, url).await?;

    if downloaded >= total_size {
        println!("  Nothing to do. File already exists.");
        return Ok(());
    }

    // Start the download
    let res = client
        .get(url)
        .header("Range", format!("bytes={}-", downloaded))
        .send()
        .await?;

    let mut stream = res.bytes_stream();

    let pb = get_progress_bar(total_size);
    pb.set_message(format!("Downloading {}", title));

    while let Some(chunk) = stream.next().await {
        let chunk = chunk?;
        let _ = file.write(&chunk)?;

        downloaded = min(downloaded + (chunk.len() as u64), total_size);
        pb.set_position(downloaded);
    }

    pb.finish_and_clear();
    println!("  Downloaded {}", title);
    Ok(())
}

fn open_file_for_write(path: &str) -> Result<(File, u64), std::io::Error> {
    if std::path::Path::new(path).exists() {
        let mut file = std::fs::OpenOptions::new()
            .read(true)
            .append(true)
            .open(path)?;

        let file_size = std::fs::metadata(path)?.len();
        file.seek(std::io::SeekFrom::Start(file_size))?;
        Ok((file, file_size))
    } else {
        let file = File::create(path)?;
        Ok((file, 0))
    }
}

async fn get_content_length(client: &Client, url: &str) -> Result<u64, DownloadError> {
    let res = client.get(url).send().await?;
    res.content_length().ok_or_else(|| {
        DownloadError::from_string(format!("Failed to get content length from '{}'", &url))
    })
}

fn get_progress_bar(total_size: u64) -> ProgressBar {
    let pb = ProgressBar::new(total_size);
    let pb_template =
        "  {msg}\n  {spinner:.green} [{elapsed}] [{bar}] {bytes} / {total_bytes} ({bytes_per_sec})";

    pb.set_style(
        ProgressStyle::default_bar()
            .template(pb_template)
            .expect("failed to parse progressbar template")
            .progress_chars("=> "),
    );
    pb
}
