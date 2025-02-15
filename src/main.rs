#[tokio::main]
async fn main() {
    tasked::logging::initialize_logging().unwrap();
    tasked::run().unwrap()
}
