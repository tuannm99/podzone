use axum::{http::Request, middleware::Next, response::Response, body::Body};

pub async fn auth_middleware(req: Request<Body>, next: Next) -> Response {
    // Example: check for a header
    if let Some(auth) = req.headers().get("x-api-key") {
        if auth == "secret" {
            return next.run(req).await;
        }
    }
    // Unauthorized
    Response::builder()
        .status(401)
        .body(Body::from("Unauthorized"))
        .unwrap()
}
