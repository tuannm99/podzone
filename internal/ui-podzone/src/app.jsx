import React from 'react';
import { Routes, Route, Navigate, Outlet } from 'react-router-dom';

import { tokenStorage } from './services/tokenStorage.js';
import PageFallback from './components/PageFallback.jsx';

const LoginPage = React.lazy(() => import('./pages/auth/LoginPage.jsx'));
const RegisterPage = React.lazy(() => import('./pages/auth/RegisterPage.jsx'));

const HomePage = React.lazy(() => import('./pages/private/HomePage.jsx'));
const SettingsPage = React.lazy(
  () => import('./pages/private/SettingsPage.jsx'),
);
const Page404 = React.lazy(() => import('./pages/Page404.jsx'));

const AppLayout = React.lazy(() => import('./layouts/layout.jsx'));

const PrivateRoute = () => {
  const isAuthenticated = !!tokenStorage.getToken();
  return isAuthenticated ? <Outlet /> : <Navigate to="/auth/login" replace />;
};

export default function App() {
  return (
    <React.Suspense fallback={<PageFallback />}>
      <Routes>
        {/* Public routes */}
        <Route path="/auth/login" element={<LoginPage />} />
        <Route path="/auth/register" element={<RegisterPage />} />

        {/* Private routes */}
        <Route element={<PrivateRoute />}>
          <Route element={<AppLayout />}>
            <Route path="/" element={<HomePage />} />
            <Route path="/home" element={<HomePage />} />
            <Route path="/settings" element={<SettingsPage />} />
          </Route>
        </Route>

        {/* Fallback */}
        <Route path="/404" element={<Page404 />} />
        <Route path="*" element={<Navigate to="/404" replace />} />
      </Routes>
    </React.Suspense>
  );
}
