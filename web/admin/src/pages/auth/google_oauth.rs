// use gloo_net::http::Request;
// use wasm_bindgen_futures::spawn_local;
use crate::services::api::api_url;
use yew::prelude::*;

#[function_component(GoogleButton)]
pub fn google_button() -> Html {
    let onclick = Callback::from(move |_| {
        let login_url = api_url("/auth/v1/google/login");
        let _ = web_sys::window().unwrap().location().set_href(&login_url);
    });

    html! {
        <button onclick={onclick} class="btn btn-outline w-full">
            <img src="https://www.svgrepo.com/show/475656/google-color.svg" class="w-5 h-5 mr-2" />
            { "Continue with Google" }
        </button>
    }
}
