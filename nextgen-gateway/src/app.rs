use crate::config::load_config;
use crate::plugins::auth::auth_middleware;
use crate::plugins::circuit_breaker::circuit_breaker_middleware;
use crate::plugins::ratelimit::ratelimit_middleware;
use crate::plugins::rewrite::rewrite_middleware;
use crate::proxy::proxy_handler;
use crate::registry::RouteRegistry;
use axum::middleware;
use axum::{Router, body::Body, http::Request, routing::any};
use tokio::net::TcpListener;
use tracing_subscriber;

pub async fn run() -> Result<(), Box<dyn std::error::Error>> {
    tracing_subscriber::fmt().init();
    let config = load_config("gateway.yaml")?;
    let registry = RouteRegistry::from_file("gateway.yaml").await?;
    let shared_registry = registry.into_shared();

    let base_app = Router::new()
        .fallback_service(any({
            let registry = shared_registry.clone();
            move |req: Request<Body>| proxy_handler(req, registry.clone())
        }))
        .merge(crate::controller::routes(shared_registry.clone()));

    let app = {
        let mut app = base_app;
        if config
            .plugins
            .as_ref()
            .and_then(|p| p.rewrite)
            .unwrap_or(false)
        {
            app = app.route_layer(middleware::from_fn(rewrite_middleware));
        }
        if config
            .plugins
            .as_ref()
            .and_then(|p| p.auth)
            .unwrap_or(false)
        {
            app = app.route_layer(middleware::from_fn(auth_middleware));
        }
        if config
            .plugins
            .as_ref()
            .and_then(|p| p.ratelimit)
            .unwrap_or(false)
        {
            app = app.route_layer(middleware::from_fn(ratelimit_middleware));
        }
        if config
            .plugins
            .as_ref()
            .and_then(|p| p.circuit_breaker)
            .unwrap_or(false)
        {
            app = app.route_layer(middleware::from_fn(circuit_breaker_middleware));
        }
        app
    };

    let listener = TcpListener::bind("0.0.0.0:3000").await?;
    axum::serve(listener, app).await?;
    Ok(())
}
