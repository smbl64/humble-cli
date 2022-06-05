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
