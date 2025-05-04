use yew::prelude::*;

use crate::layouts::Layout;
use crate::pages::auth::auth::AuthPage;

use super::router::MainRoute;

pub fn switch_root(route: MainRoute) -> Html {
    match route {
        MainRoute::Auth => html! { <AuthPage /> },
        MainRoute::NotFound => html! {<h1>{"Not Found"}</h1>},
        _ => html! { <Layout route={route} /> },
    }
}
