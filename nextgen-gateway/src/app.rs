use axum::{Router, body::Body, http::Request, middleware, routing::any};
use tokio::net::TcpListener;
use tracing_subscriber;

use crate::config::load_config;
use crate::proxy::proxy_handler;
use crate::registry::RouteRegistry;
use crate::state::AppState;

pub async fn run() -> Result<(), Box<dyn std::error::Error>> {
    tracing_subscriber::fmt().init();

    let config = load_config("gateway.yaml")?;
    let registry = RouteRegistry::from_file("gateway.yaml").await?;
    let state = AppState::new(registry.clone());

    // admin routes (CRUD /routes) - keep as RouteRegistry
    let admin = crate::controller::routes(registry.clone());

    // proxy router
    let proxy = Router::new().fallback_service(any({
        let state = state.clone();
        move |req: Request<Body>| proxy_handler(req, state.registry.clone())
    }));

    // apply plugin layers to proxy only
    let proxy = {
        let mut r = proxy;

        if config
            .plugins
            .as_ref()
            .and_then(|p| p.rewrite)
            .unwrap_or(false)
        {
            r = r.layer(middleware::from_fn(
                crate::plugins::rewrite::rewrite_middleware,
            ));
        }
        if config
            .plugins
            .as_ref()
            .and_then(|p| p.auth)
            .unwrap_or(false)
        {
            r = r.layer(middleware::from_fn(crate::plugins::auth::auth_middleware));
        }
        if config
            .plugins
            .as_ref()
            .and_then(|p| p.ratelimit)
            .unwrap_or(false)
        {
            r = r.layer(middleware::from_fn(
                crate::plugins::ratelimit::ratelimit_middleware,
            ));
        }
        if config
            .plugins
            .as_ref()
            .and_then(|p| p.circuit_breaker)
            .unwrap_or(false)
        {
            r = r.layer(middleware::from_fn(
                crate::plugins::circuit_breaker::circuit_breaker_middleware,
            ));
        }

        r
    };

    let app = admin.merge(proxy);

    let listener = TcpListener::bind("0.0.0.0:3000").await?;
    axum::serve(listener, app).await?;
    Ok(())
}
