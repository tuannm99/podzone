use axum::{
    Router,
    body::Body,
    http::{Request, Response, StatusCode},
    routing::any,
};
use reqwest::Client;
use serde::Deserialize;
use std::{collections::HashMap, fs, sync::Arc};
use tokio::net::TcpListener;
use tracing::{Level, info, warn};
use tracing_subscriber;
use url::form_urlencoded;

#[derive(Debug, Deserialize, Clone)]
struct Route {
    origin: String,
    path_prefix: Option<String>,
    path_regex: Option<String>,
    method: Option<String>,
    headers: Option<HashMap<String, String>>,
    query: Option<HashMap<String, String>>,
    rewrite: Option<String>,
    target: String,
    max_body_size: usize,
}

#[derive(Debug, Deserialize)]
struct Config {
    routes: Vec<Route>,
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt().with_max_level(Level::INFO).init();

    let config: Config =
        serde_yaml::from_str(&fs::read_to_string("gateway.yaml").expect("Missing config"))
            .expect("Invalid YAML");

    let routes = Arc::new(config.routes);
    for route in routes.iter() {
        info!(
            "Registered route: origin={} path_prefix={:?} path_regex={:?} rewrite={:?} -> target={}",
            route.origin, route.path_prefix, route.path_regex, route.rewrite, route.target
        );
    }

    let port = 3000;
    info!("Gateway running on port {}", port);

    let app = Router::new().fallback_service(any(move |req: Request<Body>| {
        let routes = routes.clone();
        async move { proxy(req, routes).await }
    }));

    let listener = TcpListener::bind(format!("0.0.0.0:{}", port))
        .await
        .unwrap();
    axum::serve(listener, app).await.unwrap();
}

async fn proxy(req: Request<Body>, routes: Arc<Vec<Route>>) -> Result<Response<Body>, StatusCode> {
    let uri = req.uri().clone();
    let method = req.method().clone();
    let host = req.headers().get("host").and_then(|v| v.to_str().ok());
    let user_agent = req
        .headers()
        .get("user-agent")
        .and_then(|v| v.to_str().ok());

    info!(
        "Incoming request: {} {} {:?} {:?}",
        method,
        uri.path(),
        host,
        user_agent
    );

    let matching_route = routes.iter().find(|route| {
        let host_match = match &route.origin[..] {
            "*" => true,
            origin => host.map_or(false, |h| h == origin),
        };

        let path_match = if let Some(prefix) = &route.path_prefix {
            uri.path().starts_with(prefix)
        } else if let Some(regex_str) = &route.path_regex {
            regex::Regex::new(regex_str)
                .map(|r| r.is_match(uri.path()))
                .unwrap_or(false)
        } else {
            false
        };

        let method_match = match &route.method {
            Some(m) => m.eq_ignore_ascii_case(method.as_str()),
            None => true,
        };

        let headers_match = match &route.headers {
            Some(hm) => hm.iter().all(|(k, v)| {
                req.headers()
                    .get(k)
                    .and_then(|val| val.to_str().ok())
                    .map_or(false, |val| val == v)
            }),
            None => true,
        };

        let query_match = match &route.query {
            Some(qm) => {
                if let Some(query) = uri.query() {
                    let parsed = form_urlencoded::parse(query.as_bytes())
                        .into_owned()
                        .collect::<HashMap<_, _>>();
                    qm.iter().all(|(k, v)| parsed.get(k) == Some(v))
                } else {
                    false
                }
            }
            None => true,
        };

        host_match && path_match && method_match && headers_match && query_match
    });

    if let Some(route) = matching_route {
        info!(
            "Matched route: origin={} -> target={}",
            route.origin, route.target
        );

        let path = uri.path_and_query().map(|x| x.as_str()).unwrap_or("/");
        let rewritten_path = if let Some(prefix) = &route.path_prefix {
            if let Some(rewrite) = &route.rewrite {
                path.strip_prefix(prefix)
                    .map(|suffix| format!("{}{}", rewrite, suffix))
                    .unwrap_or(path.to_string())
            } else {
                path.to_string()
            }
        } else {
            path.to_string()
        };

        let new_url = format!("{}{}", route.target, rewritten_path);
        let (parts, body) = req.into_parts();
        let body_bytes = axum::body::to_bytes(body, route.max_body_size)
            .await
            .map_err(|_| StatusCode::BAD_REQUEST)?;

        let client = Client::new();
        let res = client
            .request(parts.method.clone(), &new_url)
            .headers(parts.headers.clone())
            .body(body_bytes)
            .send()
            .await
            .map_err(|_| StatusCode::BAD_GATEWAY)?;

        let status = res.status();
        let headers = res.headers().clone();
        let body = res.bytes().await.unwrap_or_default();

        let mut response = Response::builder()
            .status(status)
            .body(Body::from(body))
            .unwrap();
        *response.headers_mut() = headers;
        return Ok(response);
    }

    warn!(
        "No matching route for origin={:?}, path={}",
        host,
        uri.path()
    );
    Err(StatusCode::NOT_FOUND)
}
