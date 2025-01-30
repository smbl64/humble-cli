use humble_cli::prelude::*;

fn new_download_url(web_url: &str) -> DownloadUrl {
    DownloadUrl {
        web: web_url.to_string(),
        bittorrent: web_url.to_string() + ".torrent",
    }
}

fn get_test_product() -> Product {
    let dl1 = DownloadInfo {
        md5: "".to_string(),
        format: "epub".to_string(),
        file_size: 1000,
        url: new_download_url("http://foo.com/one"),
    };

    let dl2 = DownloadInfo {
        md5: "".to_string(),
        format: "mobi".to_string(),
        file_size: 2000,
        url: new_download_url("http://foo.com/two"),
    };

    let dl_entry = ProductDownload {
        items: vec![dl1, dl2],
    };

    Product {
        machine_name: "some-book".to_string(),
        human_name: "Some Book".to_string(),
        product_details_url: "".to_string(),
        downloads: vec![dl_entry],
    }
}

#[test]
fn formats_aggregated_correctly() {
    let product = get_test_product();
    assert_eq!(product.formats_as_vec(), vec!["epub", "mobi"]);
    assert_eq!(product.formats(), "epub, mobi");
}

#[test]
fn product_total_size() {
    let product = get_test_product();
    assert_eq!(product.total_size(), 3000);
}

#[test]
fn filter_by_total_size() {
    let _product = get_test_product();
}

#[test]
fn choice_period_parses() {
    struct TestData {
        input: &'static str,
        is_ok: bool,
    }

    let data = vec![
        TestData {
            input: "invalid",
            is_ok: false,
        },
        TestData {
            input: "foobar-2023",
            is_ok: false,
        },
        TestData {
            input: "march-12000",
            is_ok: false,
        },
        TestData {
            input: "march-2023",
            is_ok: true,
        },
        TestData {
            input: "current",
            is_ok: true,
        },
    ];

    for test_data in data {
        let result = ChoicePeriod::try_from(test_data.input);
        assert_eq!(
            result.is_ok(),
            test_data.is_ok,
            "input: {}, expected ok: {}, got ok: {}",
            test_data.input,
            test_data.is_ok,
            result.is_ok()
        );
    }
}

#[test]
fn product_name_matches() {
    struct TestData {
        name: String,
        keywords: String,
        match_mode: MatchMode,
        expected: bool,
    }

    let test_data = vec![
        TestData {
            name: "Python programming".to_owned(),
            keywords: "python".to_owned(),
            match_mode: MatchMode::Any,
            expected: true,
        },
        TestData {
            name: "Python programming".to_owned(),
            keywords: "java".to_owned(),
            match_mode: MatchMode::Any,
            expected: false,
        },
        TestData {
            name: "Programming in Rust second edition".to_owned(),
            keywords: "rust edition".to_owned(),
            match_mode: MatchMode::All,
            expected: true,
        },
        TestData {
            name: "Programming in Rust second edition".to_owned(),
            keywords: "rust third".to_owned(),
            match_mode: MatchMode::All,
            expected: false,
        },
    ];

    for td in test_data {
        let mut product = Product::default();
        product.human_name = td.name.clone();

        let keywords = td.keywords.to_lowercase();
        let keywords: Vec<&str> = keywords.split(" ").collect();
        let got = product.name_matches(&keywords, &td.match_mode);
        assert_eq!(
            got, td.expected,
            "expected {}, got {}, name = {}, keywords = {}, mode = {:?}",
            td.expected, got, td.name, td.keywords, &td.match_mode,
        )
    }
}
