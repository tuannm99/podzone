#![recursion_limit = "1024"]
mod components;
mod layouts;
mod pages;
mod routes;

use routes::{MainRoute, switch_root};
use yew::prelude::*;
use yew_router::prelude::*;

#[function_component(App)]
fn app() -> Html {
    html! {
        <BrowserRouter>
            <Switch<MainRoute> render={switch_root} />
        </BrowserRouter>
    }
}

fn main() {
    yew::Renderer::<App>::new().render();
}
