use byte_unit::Byte;
use std::{collections::HashSet, future::Future};

pub fn run_future<F, T>(input: F) -> T
where
    F: Future<Output = T>,
{
    let runtime = tokio::runtime::Runtime::new().unwrap();
    runtime.block_on(input)
}

pub fn humanize_bytes(bytes: u64) -> String {
    Byte::from_bytes(bytes)
        .get_appropriate_unit(true)
        .to_string()
}

// Convert a string representing a byte size (e.g. 12MB) to a number.
// It supports the IEC (KiB MiB ...) and KB MB ... formats.
pub fn byte_string_to_number(byte_string: &str) -> Option<u64> {
    Byte::from_str(byte_string).map(|b| b.into()).ok()
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

pub fn str_vectors_intersect<T1, T2>(first: &[T1], second: &[T2]) -> bool
where
    T1: AsRef<str>,
    T2: AsRef<str>,
{
    if first.is_empty() || second.is_empty() {
        return false;
    }

    let mut first_set = HashSet::new();

    for first_item in first {
        first_set.insert(first_item.as_ref().to_lowercase());
    }

    for second_item in second {
        if first_set.contains(&second_item.as_ref().to_lowercase()) {
            return true;
        }
    }

    false
}

#[test]
fn test_remove_invalid_chars() {
    assert_eq!(
        replace_invalid_chars_in_filename("Humble Bundle: Nice book"),
        "Humble Bundle  Nice book".to_string()
    );
}

#[test]
fn test_extract_filename_from_url() {
    let test_data = vec![(
        "with filename",
        "https://dl.humble.com/grokkingalgorithms.mobi?gamekey=xxxxxx&ttl=1655031034&t=yyyyyyyyyy",
        Some("grokkingalgorithms.mobi".to_string()),
    ), (
        "no filename",
        "https://www.google.com/",
        None
    )];

    for (name, url, expected) in test_data {
        assert_eq!(
            extract_filename_from_url(url),
            expected,
            "test case '{}'",
            name
        );
    }
}

#[test]
fn vecor_inter() {
    let test_data = vec![
        (vec!["FOO", "bar"], vec!["foo"], true),
        (vec!["foo", "bar"], vec!["baz"], false),
        (vec!["foo"], vec![], false),
        (vec![], vec!["baz"], false),
    ];

    for (first, second, result) in test_data {
        let msg = format!(
            "intersect of {:?} and {:?}, expected: {}",
            first, second, result
        );
        assert_eq!(str_vectors_intersect(&first, &second), result, "{}", msg);
    }
}
