pub mod args;

use clap::Parser;
use tasked::args::Args;
use tracing::info;

#[tokio::main]
async fn main() {
    tasked::tracer::initialize_logging().unwrap();
    info!("Application Started");
    let args = Args::parse();
    tasked::run(args).await.unwrap()
}
