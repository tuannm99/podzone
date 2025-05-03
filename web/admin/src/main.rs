#![recursion_limit = "1024"]

use yew::prelude::*;
use yew_router::prelude::*;

use self::routes::{MainRoute, switch_main};

mod components;
mod pages;
mod routes;

#[function_component(App)]
fn app() -> Html {
    yew::html! {
        <BrowserRouter>
            <Switch<MainRoute> render={switch_main} />
        </BrowserRouter>
    }
}

fn main() {
    yew::Renderer::<App>::new().render();
}
