export const shellRouteComponents = {
  acceptInvite: () => import('./pages/auth/AcceptInvitePage'),
  devAuthBootstrap: () => import('./pages/auth/DevAuthBootstrapPage'),
  googleCallback: () => import('./pages/auth/GoogleCallbackPage'),
  login: () => import('./pages/auth/LoginPage'),
  register: () => import('./pages/auth/RegisterPage'),
};
