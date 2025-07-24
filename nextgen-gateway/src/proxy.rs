use crate::registry::RouteRegistry;
use axum::{
    body::Body,
    http::{Request, Response, StatusCode},
};
use regex::Regex;
use reqwest::Client;
use std::collections::HashMap;
use tracing::warn;
use url::form_urlencoded;

pub async fn proxy_handler(
    req: Request<Body>,
    registry: RouteRegistry,
) -> Result<Response<Body>, StatusCode> {
    let uri = req.uri().clone();
    let method = req.method().clone();
    let host = req.headers().get("host").and_then(|v| v.to_str().ok());

    let routes = registry.get_routes().await;
    let matched = routes.iter().find(|route| {
        let host_match = match route.origin.as_str() {
            "*" => true,
            origin => host.map_or(false, |h| h == origin),
        };
        let path_match = route
            .path_prefix
            .as_ref()
            .map_or(false, |prefix| uri.path().starts_with(prefix))
            || route
                .path_regex
                .as_ref()
                .and_then(|r| Regex::new(r).ok())
                .map_or(false, |re| re.is_match(uri.path()));
        let method_match = route
            .method
            .as_ref()
            .map_or(true, |m| m.eq_ignore_ascii_case(method.as_str()));
        let headers_match = route.headers.as_ref().map_or(true, |hm| {
            hm.iter().all(|(k, v)| {
                req.headers()
                    .get(k)
                    .and_then(|val| val.to_str().ok())
                    .map_or(false, |val| val == v)
            })
        });
        let query_match = route.query.as_ref().map_or(true, |qm| {
            uri.query().map_or(false, |q| {
                let parsed: HashMap<_, _> =
                    form_urlencoded::parse(q.as_bytes()).into_owned().collect();
                qm.iter().all(|(k, v)| parsed.get(k) == Some(v))
            })
        });

        host_match && path_match && method_match && headers_match && query_match
    });

    if let Some(route) = matched {
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
        let body = res.bytes().await.map_err(|_| StatusCode::BAD_GATEWAY)?;

        let mut response = Response::builder().status(status);
        // Copy only valid headers
        for (key, value) in headers.iter() {
            if let Ok(val) = value.to_str() {
                response = response.header(key, val);
            }
        }
        let response = response
            .body(Body::from(body))
            .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;
        Ok(response)
    } else {
        warn!("No matching route for host={:?}, path={}", host, uri.path());
        Err(StatusCode::NOT_FOUND)
    }
}
