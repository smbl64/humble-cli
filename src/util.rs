use std::future::Future;

pub fn run_future<F, T>(input: F) -> T
where
    F: Future<Output = T>,
{
    let runtime = tokio::runtime::Runtime::new().unwrap();
    runtime.block_on(input)
}

pub fn humanize_bytes(bytes: u64) -> String {
    bytesize::to_string(bytes, true)
}

pub fn replace_invalid_chars_in_filename(input: &str) -> String {
    let replacement: char = ' ';
    let invalid_chars: Vec<char> =
        vec!['/', '\\', '?', '%', '*', ':', '|', '"', '<', '>', ';', '='];

    input
        .chars()
        .map(|c| {
            if invalid_chars.contains(&c) {
                replacement
            } else {
                c
            }
        })
        .collect::<String>()
}

pub fn extract_filename_from_url(url: &str) -> Option<String> {
    let url = reqwest::Url::parse(url);
    if url.is_err() {
        return None;
    }

    let url = url.unwrap();
    let path_segments: Vec<&str> = url.path_segments()?.collect();
    match path_segments[path_segments.len() - 1] {
        "" => None,
        segment => Some(segment.to_string()),
    }
}

#[test]
fn test_remove_invalid_chars() {
    assert_eq!(
        replace_invalid_chars_in_filename("Humble Bundle: Book = Nice"),
        "Humble Bundle_ Book _ Nice".to_string()
    );
}

#[test]
fn test_extract_filename_from_url() {
    let test_data = vec![(
        "with filename",
        "https://dl.humble.com/grokkingalgorithms.mobi?gamekey=xxxxxx&ttl=1655031034&t=yyyyyyyyyy",
        Some("grokkingalgorithms.mobi".to_string()),
    ),
    ( 
        "no filename",
        "https://www.google.com/", None
    )
    ];

    for (name, url, expected) in test_data {
        assert_eq!(
            extract_filename_from_url(url),
            expected,
            "test case '{}'",
            name
        );
    }
}
