use yew::prelude::*;
use yew_router::prelude::*;

use crate::pages::auth::AuthPage;

#[derive(Clone, Routable, PartialEq)]
pub enum MainRoute {
    #[at("/")]
    Home,
    #[at("/auth")]
    Auth,

    #[at("/settings")]
    SettingsRoot,
    #[at("/settings/*")]
    Settings,

    #[not_found]
    #[at("/404")]
    NotFound,
}

#[derive(Clone, Routable, PartialEq)]
pub enum SettingsRoute {
    #[at("/settings")]
    Profile,
    #[at("/settings/friends")]
    Friends,
    #[at("/settings/theme")]
    Theme,
    #[not_found]
    #[at("/settings/404")]
    NotFound,
}

pub fn switch_main(route: MainRoute) -> Html {
    match route {
        MainRoute::Home => html! {<h1>{"Home"}</h1>},
        MainRoute::Auth => html! {<AuthPage />},
        MainRoute::SettingsRoot | MainRoute::Settings => {
            html! { <Switch<SettingsRoute> render={switch_settings} /> }
        }
        MainRoute::NotFound => html! {<h1>{"Not Found"}</h1>},
    }
}

pub fn switch_settings(route: SettingsRoute) -> Html {
    match route {
        SettingsRoute::Profile => html! {<h1>{"Profile"}</h1>},
        SettingsRoute::Friends => html! {<h1>{"Friends"}</h1>},
        SettingsRoute::Theme => html! {<h1>{"Theme"}</h1>},
        SettingsRoute::NotFound => html! {<Redirect<MainRoute> to={MainRoute::NotFound}/>},
    }
}
