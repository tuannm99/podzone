use crate::config::Route;
use crate::registry::RouteRegistry;
use axum::{
    Router,
    extract::{Json, Path, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{delete, get, post, put},
};

pub fn routes(registry: RouteRegistry) -> Router {
    Router::new()
        .route("/routes", get(get_all_routes))
        .route("/routes", post(set_all_routes))
        .route("/routes/{idx}", put(update_route))
        .route("/routes/{idx}", delete(delete_route))
        .with_state(registry)
}

async fn get_all_routes(
    registry: State<RouteRegistry>,
) -> axum::response::Response {
    let routes = (registry.0).0.read().await;
    Json(routes.clone()).into_response()
}

async fn set_all_routes(
    registry: State<RouteRegistry>,
    Json(new_routes): Json<Vec<Route>>,
) -> axum::response::Response {
    let mut routes = (registry.0).0.write().await;
    *routes = new_routes;
    Json("ok").into_response()
}

async fn update_route(
    registry: State<RouteRegistry>,
    Path(idx): Path<usize>,
    Json(route): Json<Route>,
) -> (StatusCode, Json<&'static str>) {
    let mut routes = (registry.0).0.write().await;
    if idx < routes.len() {
        routes[idx] = route;
        (StatusCode::OK, Json("updated"))
    } else {
        (StatusCode::NOT_FOUND, Json("index out of bounds"))
    }
}

async fn delete_route(
    registry: State<RouteRegistry>,
    Path(idx): Path<usize>,
) -> (StatusCode, Json<&'static str>) {
    let mut routes = (registry.0).0.write().await;
    if idx < routes.len() {
        routes.remove(idx);
        (StatusCode::OK, Json("deleted"))
    } else {
        (StatusCode::NOT_FOUND, Json("index out of bounds"))
    }
}
