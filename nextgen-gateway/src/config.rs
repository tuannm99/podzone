use serde::{Deserialize, Serialize};

#[derive(Debug, Deserialize, Serialize, Clone)]
pub struct Route {
    pub origin: String,
    pub path_prefix: Option<String>,
    pub path_regex: Option<String>,
    pub method: Option<String>,
    pub headers: Option<std::collections::HashMap<String, String>>,
    pub query: Option<std::collections::HashMap<String, String>>,
    pub rewrite: Option<String>,
    pub target: String,
    pub max_body_size: usize,
}

#[derive(Debug, Deserialize, Serialize, Clone, Default)]
pub struct PluginsConfig {
    pub auth: Option<bool>,
    pub ratelimit: Option<bool>,
    pub circuit_breaker: Option<bool>,
    pub rewrite: Option<bool>,
}

#[derive(Debug, Deserialize)]
pub struct Config {
    pub routes: Vec<Route>,
    pub plugins: Option<PluginsConfig>,
}

pub fn load_config(path: &str) -> Result<Config, Box<dyn std::error::Error>> {
    let content = std::fs::read_to_string(path)?;
    let config = serde_yaml::from_str(&content)?;
    Ok(config)
}
