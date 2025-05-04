use crate::layouts::{Header, Sidebar};
use crate::routes::router::{switch_main, MainRoute};
use yew::prelude::*;

#[derive(Properties, PartialEq)]
pub struct LayoutProps {
    pub route: MainRoute,
}

#[function_component(Layout)]
pub fn layout(props: &LayoutProps) -> Html {
    html! {
        <div class="flex flex-col h-screen">
            <Header />
            <div class="flex flex-1 overflow-hidden">
                <Sidebar />
                <main class="flex-1 overflow-y-auto p-6">
                    { switch_main(props.route.clone()) }
                </main>
            </div>
        </div>
    }
}
