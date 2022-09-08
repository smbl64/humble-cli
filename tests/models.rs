use humble_cli::humble_api::*;

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

    let product = Product {
        machine_name: "some-book".to_string(),
        human_name: "Some Book".to_string(),
        product_details_url: "".to_string(),
        downloads: vec![dl_entry],
    };

    product
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
