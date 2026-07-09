export const AUTH_LOGIN_REQUIRED_EVENT = 'podzone:auth-login-required'

export function requestLoginNavigation() {
    if (typeof window === 'undefined') return
    window.dispatchEvent(new CustomEvent(AUTH_LOGIN_REQUIRED_EVENT))
}
