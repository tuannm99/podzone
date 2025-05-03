use yew::prelude::{Html, function_component, html};

#[function_component(LoginForm)]
pub fn login_form() -> Html {
    html! {
        <form class="space-y-4">
            <input type="email" placeholder="Email" class="input input-bordered w-full" />
            <input type="password" placeholder="Password" class="input input-bordered w-full" />
            <button type="submit" class="btn btn-primary w-full">{"Login"}</button>
        </form>
    }
}
