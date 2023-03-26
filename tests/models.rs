use humble_cli::prelude::*;

fn new_download_url(web_url: &str) -> DownloadUrl {
    DownloadUrl {
        web: web_url.to_string(),
        bittorrent: "".to_string(),
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
