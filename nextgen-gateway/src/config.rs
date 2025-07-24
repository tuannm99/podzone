use serde::Deserialize;
use std::fs;

#[derive(Debug, Deserialize, Clone)]
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

#[derive(Debug, Deserialize)]
pub struct Config {
    pub routes: Vec<Route>,
}

pub fn load_config(path: &str) -> Config {
    let content = fs::read_to_string(path).expect("Missing config");
    serde_yaml::from_str(&content).expect("Invalid YAML")
}
