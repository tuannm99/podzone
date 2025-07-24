use crate::proxy::proxy_handler;
use crate::registry::RouteRegistry;
use axum::{Router, body::Body, http::Request, routing::any};
use tokio::net::TcpListener;
use tracing_subscriber;

pub async fn run() -> Result<(), Box<dyn std::error::Error>> {
    tracing_subscriber::fmt().init();
    let registry = RouteRegistry::from_file("gateway.yaml").await?;
    let shared_registry = registry.into_shared();

    let app = Router::new()
        .fallback_service(any({
            let registry = shared_registry.clone();
            move |req: Request<Body>| proxy_handler(req, registry.clone())
        }))
        .merge(crate::controller::routes(shared_registry.clone()));

    let listener = TcpListener::bind("0.0.0.0:3000").await?;
    axum::serve(listener, app).await?;
    Ok(())
}
