use yew::prelude::*;

#[function_component(Header)]
pub fn header() -> Html {
    html! {
        <header class="navbar bg-base-100 shadow-md">
            <div class="flex-1 text-xl font-bold px-4">{"Admin App"}</div>
            <div class="flex-none gap-2 px-4">
                <button class="btn btn-sm">{"Logout"}</button>
            </div>
        </header>
    }
}
