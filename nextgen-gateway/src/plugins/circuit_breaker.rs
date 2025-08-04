use axum::{http::Request, middleware::Next, response::Response, body::Body};
use std::sync::Mutex;
use std::time::{Duration, Instant};
use once_cell::sync::Lazy;

static STATE: Lazy<Mutex<CircuitState>> = Lazy::new(|| Mutex::new(CircuitState::default()));

#[derive(Default)]
struct CircuitState {
    failures: u32,
    open_until: Option<Instant>,
}

const FAILURE_THRESHOLD: u32 = 5;
const OPEN_DURATION: Duration = Duration::from_secs(30);

pub async fn circuit_breaker_middleware(req: Request<Body>, next: Next) -> Response {
    let mut state = STATE.lock().unwrap();
    
    if let Some(until) = state.open_until {
        if Instant::now() < until {
            return Response::builder()
                .status(503)
                .body(Body::from("Circuit open - try again later"))
                .unwrap();
        } else {
            state.failures = 0;
            state.open_until = None;
        }
    }

    drop(state);

    let response = next.run(req).await;
    let status = response.status();
    let mut state = STATE.lock().unwrap();

    if status.is_server_error() {
        state.failures += 1;
        if state.failures >= FAILURE_THRESHOLD {
            state.open_until = Some(Instant::now() + OPEN_DURATION);
        }
    } else {
        state.failures = 0;
    }
    
    response
}
