use crate::config::{Config, Route, load_config};
use std::sync::Arc;
use tokio::sync::RwLock;

#[derive(Clone)]
pub struct RouteRegistry(pub Arc<RwLock<Vec<Route>>>);

impl RouteRegistry {
    pub async fn from_file(path: &str) -> Self {
        let config = load_config(path);
        Self(Arc::new(RwLock::new(config.routes)))
    }

    pub fn into_shared(self) -> Self {
        self
    }

    pub async fn get_routes(&self) -> Vec<Route> {
        self.0.read().await.clone()
    }

    pub async fn set_routes(&self, routes: Vec<Route>) {
        *self.0.write().await = routes;
    }
}
