use yew::prelude::{Callback, Html, function_component, html, use_state};

use login::LoginForm;
use register::RegisterForm;

use super::login;
use super::register;

#[function_component(AuthPage)]
pub fn auth_page() -> Html {
    let is_login = use_state(|| true);
    let toggle = {
        let is_login = is_login.clone();
        Callback::from(move |_| is_login.set(!*is_login))
    };

    html! {
        <div class="flex items-center justify-center min-h-screen bg-base-200">
            <div class="card w-full max-w-md bg-base-100 shadow-xl p-8">
                <h2 class="text-2xl font-bold text-center mb-6">
                    { if *is_login { "Login" } else { "Register" } }
                </h2>

                {
                    if *is_login {
                        html! { <LoginForm /> }
                    } else {
                        html! { <RegisterForm /> }
                    }
                }

                <div class="divider">{"OR"}</div>

                <button onclick={|_| web_sys::window().unwrap().location().set_href("/auth/google").unwrap()} class="btn btn-outline w-full">
                    <img src="https://www.svgrepo.com/show/475656/google-color.svg" class="w-5 h-5 mr-2" />
                    { "Continue with Google" }
                </button>

                <p class="text-center mt-4 text-sm">
                    {
                        if *is_login {
                            html! {
                                <>
                                    {"Don't have an account? "}
                                    <button onclick={toggle.clone()} class="link link-primary">{"Register"}</button>
                                </>
                            }
                        } else {
                            html! {
                                <>
                                    {"Already have an account? "}
                                    <button onclick={toggle.clone()} class="link link-primary">{"Login"}</button>
                                </>
                            }
                        }
                    }
                </p>
            </div>
        </div>
    }
}

