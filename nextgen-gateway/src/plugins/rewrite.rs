use axum::{http::{Request, Uri}, middleware::Next, response::Response, body::Body};

pub async fn rewrite_middleware(mut req: Request<Body>, next: Next) -> Response {
    let orig_path = req.uri().path();
    if let Some(stripped) = orig_path.strip_prefix("/api/v1") {
        let new_path = format!("/v1{}", stripped);
        let mut parts = req.uri().clone().into_parts();
        parts.path_and_query = Some(new_path.parse().unwrap());
        *req.uri_mut() = Uri::from_parts(parts).unwrap();
    }
    next.run(req).await
}
