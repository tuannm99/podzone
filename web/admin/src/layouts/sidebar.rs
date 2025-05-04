use yew::prelude::*;
use yew_router::prelude::*;

use crate::routes::router::{MainRoute, SettingsRoute};

#[function_component(Sidebar)]
pub fn sidebar() -> Html {
    html! {
        <aside class="w-64 bg-base-200 p-4 space-y-2">
            <nav class="menu bg-base-200 rounded-box w-56">
                { link_button("Home", MainRoute::Home) }
                { link_button("Settings", MainRoute::SettingsRoot) }
                <li>
                    <Link<SettingsRoute> to={SettingsRoute::NotFound} classes="btn btn-ghost w-full justify-start">
                        { "Settings404" }
                    </Link<SettingsRoute>>
                </li>
            </nav>
        </aside>
    }
}

fn link_button(label: &str, route: MainRoute) -> Html {
    html! {
        <li>
            <Link<MainRoute> to={route} classes="btn btn-ghost w-full justify-start">
                { label }
            </Link<MainRoute>>
        </li>
    }
}
