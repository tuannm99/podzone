use crate::config::Route;
use crate::registry::RouteRegistry;
use axum::{
    Router,
    extract::{Json, Path},
    routing::{delete, get, post, put},
};

pub fn routes(registry: RouteRegistry) -> Router {
    Router::new()
        .route("/routes", get(get_all_routes))
        .route("/routes", post(set_all_routes))
        .route("/routes/:idx", put(update_route))
        .route("/routes/:idx", delete(delete_route))
        .with_state(registry)
}

async fn get_all_routes(
    axum::extract::State(registry): axum::extract::State<RouteRegistry>,
) -> Json<Vec<Route>> {
    Json(registry.get_routes().await)
}

async fn set_all_routes(
    axum::extract::State(registry): axum::extract::State<RouteRegistry>,
    Json(new_routes): Json<Vec<Route>>,
) -> Json<&'static str> {
    registry.set_routes(new_routes).await;
    Json("ok")
}

async fn update_route(
    axum::extract::State(registry): axum::extract::State<RouteRegistry>,
    Path(idx): Path<usize>,
    Json(route): Json<Route>,
) -> Json<&'static str> {
    let mut routes = registry.0.write().await;
    if idx < routes.len() {
        routes[idx] = route;
        Json("updated")
    } else {
        Json("index out of bounds")
    }
}

async fn delete_route(
    axum::extract::State(registry): axum::extract::State<RouteRegistry>,
    Path(idx): Path<usize>,
) -> Json<&'static str> {
    let mut routes = registry.0.write().await;
    if idx < routes.len() {
        routes.remove(idx);
        Json("deleted")
    } else {
        Json("index out of bounds")
    }
}
