use yew::prelude::{Callback, Html, function_component, html, use_state};

use super::google_oauth::GoogleButton;
use super::login::LoginForm;
use super::register::RegisterForm;

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

                <GoogleButton />

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
