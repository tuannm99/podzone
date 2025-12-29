use axum::{
    body::Body,
    http::{Request, StatusCode},
    middleware::Next,
    response::{IntoResponse, Response},
};

use once_cell::sync::Lazy;
use std::collections::HashMap;
use std::sync::Mutex;
use std::time::{Duration, Instant};

static RATE_LIMITER: Lazy<Mutex<HashMap<String, (u32, Instant)>>> =
    Lazy::new(|| Mutex::new(HashMap::new()));
const LIMIT: u32 = 10;
const WINDOW: Duration = Duration::from_secs(60);

pub async fn ratelimit_middleware(req: Request<Body>, next: Next) -> Response {
    let ip = req
        .headers()
        .get("x-real-ip")
        .and_then(|v| v.to_str().ok())
        .unwrap_or("unknown")
        .to_string();

    let now = Instant::now();
    {
        let mut map = RATE_LIMITER.lock().unwrap();
        let entry = map.entry(ip).or_insert((0, now));

        if now.duration_since(entry.1) > WINDOW {
            *entry = (1, now);
        } else {
            entry.0 += 1;
            if entry.0 > LIMIT {
                return (StatusCode::TOO_MANY_REQUESTS, "Too Many Requests").into_response();
            }
        }
    }

    next.run(req).await
}
