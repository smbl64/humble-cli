use byte_unit::{Byte, UnitType};
use std::{collections::HashSet, future::Future};

pub fn run_future<F, T>(input: F) -> T
where
    F: Future<Output = T>,
{
    let runtime = tokio::runtime::Runtime::new().unwrap();
    runtime.block_on(input)
}

pub fn humanize_bytes(bytes: u64) -> String {
    let b = Byte::from_u64(bytes).get_appropriate_unit(UnitType::Binary);
    format!("{b:.2}")
}

// Convert a string representing a byte size (e.g. 12MB) to a number.
// It supports the IEC (KiB MiB ...) and KB MB ... formats.
pub fn byte_string_to_number(byte_string: &str) -> Option<u64> {
    Byte::parse_str(byte_string, true).map(|b| b.into()).ok()
}

pub fn replace_invalid_chars_in_filename(input: &str) -> String {
    let replacement: char = ' ';
    let invalid_chars: Vec<char> = vec![
        '/', '\\', '?', '%', '*', ':', '|', '"', '<', '>', ';', '=', '\n',
    ];

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
        .trim()
        .to_string()
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
    let dash_idx = value.find('-');

    if dash_idx.is_none() {
        return value.parse::<usize>().map(|v| vec![v]).ok();
    }

    let dash_idx = dash_idx.unwrap();

    let left = &value[0..dash_idx];
    let right = &value[dash_idx + 1..];

    let range_left = if !left.is_empty() {
        match left.parse::<usize>() {
            Ok(v) => v,
            Err(_) => return None,
        }
    } else {
        1
    };

    let range_right = if !right.is_empty() {
        match right.parse::<usize>() {
            Ok(v) => v,
            Err(_) => return None,
        }
    } else {
        max_value
    };

    // These min and max values are intentional:
    // min value is `1` and max value is `max_value + 1`
    Some((range_left..range_right + 1).collect())
}

pub fn union_usize_ranges(values: &[&str], max_value: usize) -> Result<Vec<usize>, anyhow::Error> {
    let mut invalid_values = vec![];
    let mut parsed = HashSet::new();

    for &v in values {
        match parse_usize_range(v, max_value) {
            Some(usize_values) => parsed.extend(usize_values),
            None => invalid_values.push(v),
        }
    }

    if !invalid_values.is_empty() {
        let msg = invalid_values
            .into_iter()
            .map(|v| format!("'{}'", v))
            .collect::<Vec<_>>()
            .join(", ");

        return Err(anyhow::anyhow!("{}", msg));
    }

    let mut output = Vec::from_iter(parsed);
    output.sort();
    Ok(output)
}

#[test]
fn test_remove_invalid_chars() {
    let test_data = vec![
        ("Humble Bundle: Nice book", "Humble Bundle  Nice book"),
        ("::Make::", "Make"),
        ("Test\nFile", "Test File"),
    ];

    for (input, expected) in test_data {
        let got = replace_invalid_chars_in_filename(input);
        assert_eq!(expected, got);
    }
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
/// A test to make sure `humanize_bytes` function works as expected.
///
/// We rely on an external library to do this for us, but still a good
/// idea to have a small test to make sure the library is not broken :-)
fn test_humanize_bytes() {
    let test_data = vec![(1, "1 B"), (3 * 1024, "3.00 KiB")];

    for (input, want) in test_data {
        assert_eq!(humanize_bytes(input), want.to_string());
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
        ("invalid start", "abc-", None),
        ("invalid end", "-abc", None),
        ("invalid start and end", "abc-def", None),
    ];

    for (name, input, expected) in test_data {
        let msg = format!(
            "'{}' failed: input = {}, expected = {:?}",
            name, input, &expected
        );
        assert_eq!(parse_usize_range(input, MAX_VAL), expected, "{}", msg);
    }
}

#[test]
fn test_union_valid_usize_ranges() {
    const MAX_VAL: usize = 10;
    let test_data = vec![
        ("simple values", vec!["5", "10"], vec![5, 10]),
        ("simple value and range", vec!["8", "7-"], vec![7, 8, 9, 10]),
        ("two ranges", vec!["-3", "7-"], vec![1, 2, 3, 7, 8, 9, 10]),
    ];

    for (name, input, expected) in test_data {
        let output = union_usize_ranges(&input, MAX_VAL);

        let msg = format!(
            "'{}' failed: input = {:?}, expected = {:?}",
            name, &input, &expected
        );

        assert!(output.is_ok(), "{}", msg);
        assert_eq!(output.unwrap(), expected, "{}", msg);
    }
}

#[test]
fn test_union_invalid_usize_ranges() {
    const MAX_VAL: usize = 10;
    let test_data = vec![
        ("invalid simple values", vec!["a", "b"]),
        ("invalid ranges", vec!["a-", "-b"]),
    ];

    for (name, input) in test_data {
        // expected error message
        let expected_err_msg = input
            .iter()
            .map(|v| format!("'{}'", v))
            .collect::<Vec<_>>()
            .join(", ");

        let output = union_usize_ranges(&input, MAX_VAL);

        let assert_msg = format!(
            "'{}' failed: input = {:?}, expected = {:?}",
            name, &input, &expected_err_msg
        );

        assert!(output.is_err(), "{}", assert_msg);
        let output_err_msg: String = output.unwrap_err().downcast().unwrap();
        assert_eq!(output_err_msg, expected_err_msg, "{}", assert_msg);
    }
}

#[cfg(target_os = "windows")]
#[test]
fn test_windows_filename_validation() {
    use std::fs::File;

    // These are actual forbidden characters on Windows; they are a subset of what
    // `replace_invalid_chars_in_filename` replaces.
    let invalid_chars = vec!['/', '\\', '?', '*', ':', '|', '"', '<', '>', '\n'];

    for &c in &invalid_chars {
        let filename = format!("test_{}_file.txt", c);
        // Attempt to create file with invalid character
        let result = File::create(&filename);
        assert!(
            result.is_err(),
            "Expected error creating file with invalid character '{}'",
            c
        );

        // Replace invalid characters
        let cleaned = replace_invalid_chars_in_filename(&filename);
        assert_ne!(
            cleaned, filename,
            "Filename not cleaned for character '{}'",
            c
        );

        // Attempt to create file with cleaned name
        File::create(&cleaned).expect(&format!(
            "Failed to create cleaned file for character '{}'",
            c
        ));
        // Clean up the file after test
        std::fs::remove_file(&cleaned).expect("Failed to remove test file");
    }
}
