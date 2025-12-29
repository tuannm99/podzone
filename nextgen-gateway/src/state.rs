use crate::registry::RouteRegistry;
use hyper_util::client::legacy::{Client, connect::HttpConnector};
use hyper_util::rt::TokioExecutor;

#[derive(Clone)]
pub struct AppState {
    pub registry: RouteRegistry,
    pub client: Client<HttpConnector, axum::body::Body>,
}

impl AppState {
    pub fn new(registry: RouteRegistry) -> Self {
        // Build a Hyper client that supports HTTP/1 and HTTP/2 (good for gRPC).
        let mut connector = HttpConnector::new();
        connector.enforce_http(false);

        let client = Client::builder(TokioExecutor::new())
            .http2_adaptive_window(true)
            .build(connector);

        Self { registry, client }
    }
}
