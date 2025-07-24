mod app;
mod config;
mod controller;
mod proxy;
mod registry;

pub use routes::routes;

#[tokio::main]
async fn main() {
    app::run().await;
}
