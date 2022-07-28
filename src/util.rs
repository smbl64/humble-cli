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

/// Parse the given `usize` range and return the values in that range as a `Vector`.
///
/// Value formats are:
/// - A single value: 42
/// - A range with beginning and end (1-5): Returns all valus between those two numbers (inclusive).
/// - A range with no end (10-): In this case, `max_value` specifies the end of the range.
/// - A range with no beginning (-5): In this case, the range begins at `1`.
///
/// Note: the range starts at `1`, **not** `0`.
pub fn parse_usize_range(value: &str, max_value: usize) -> Option<Vec<usize>> {
    let dash_idx = value.find("-");

    if dash_idx == None {
        return value.parse::<usize>().map(|v| vec![v]).ok();
    }

    let dash_idx = dash_idx.unwrap();

    let left = &value[0..dash_idx];
    let right = &value[dash_idx + 1..];

    let range_left = left.parse::<usize>().unwrap_or(1);
    let range_right = right.parse::<usize>().unwrap_or(max_value);

    // These min and max values are intentional:
    // min value is `1` and max value is `max_value + 1`
    Some((range_left..range_right + 1).collect())
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
fn test_vectors_intersect() {
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

#[test]
fn test_parse_usize_range() {
    const MAX_VAL: usize = 50;

    let test_data = vec![
        ("empty string", "", None),
        ("invalid string", "abcd", None),
        ("single value", "42", Some(vec![42])),
        (
            "range with start and end",
            "5-10",
            Some(vec![5, 6, 7, 8, 9, 10]),
        ),
        ("range with no start", "-5", Some(vec![1, 2, 3, 4, 5])),
        (
            "range with no end",
            "45-",
            Some(vec![45, 46, 47, 48, 49, 50]),
        ), // 50 is MAX_VAL
    ];

    for (name, input, expected) in test_data {
        let msg = format!(
            "'{}' failed: input = {}, expected = {:?}",
            name, input, &expected
        );
        assert_eq!(parse_usize_range(input, MAX_VAL), expected, "{}", msg);
    }
}
