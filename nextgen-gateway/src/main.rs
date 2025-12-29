mod app;
mod config;
mod controller;
mod plugins;
mod proxy;
mod registry;
mod state;

#[tokio::main]
async fn main() {
    if let Err(e) = app::run().await {
        eprintln!("Application error: {}", e);
        std::process::exit(1);
    }
}
