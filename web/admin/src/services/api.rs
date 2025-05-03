pub fn api_url(path: &str) -> String {
    let base = option_env!("BASE_URL").unwrap_or("http://localhost:8080");
    format!("{base}{path}")
}

