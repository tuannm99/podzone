import React from 'react';
import { Routes, Route, Navigate, Outlet } from 'react-router-dom';
import { tokenStorage } from './services/tokenStorage';
import PageFallback from './components/PageFallback';

const LoginPage = React.lazy(() => import('./pages/auth/LoginPage.jsx'));
const RegisterPage = React.lazy(() => import('./pages/auth/RegisterPage.jsx'));
const Page404 = React.lazy(() => import('./pages/Page404.jsx'));

// Layouts
const AdminLayout = React.lazy(() => import('./layouts/AdminLayout.jsx'));
const TenantLayout = React.lazy(() => import('./layouts/TenantLayout.jsx'));

// Admin pages (REST)
const AdminHome = React.lazy(() => import('./pages/admin/AdminHome.jsx'));
const AdminSettings = React.lazy(
  () => import('./pages/admin/AdminSettings.jsx'),
);

// Tenant pages (GraphQL)
const TenantHome = React.lazy(() => import('./pages/tenant/TenantHome.jsx'));
const TenantOrders = React.lazy(
  () => import('./pages/tenant/TenantOrders.jsx'),
);

const RequireAuth = () => {
  const ok = !!tokenStorage.getToken();
  return ok ? <Outlet /> : <Navigate to="/auth/login" replace />;
};

export default function App() {
  return (
    <React.Suspense fallback={<PageFallback />}>
      <Routes>
        {/* Public */}
        <Route path="/auth/login" element={<LoginPage />} />
        <Route path="/auth/register" element={<RegisterPage />} />
        <Route path="/404" element={<Page404 />} />

        {/* Protected */}
        <Route element={<RequireAuth />}>
          {/* Admin (REST) */}
          <Route element={<AdminLayout />}>
            <Route path="/admin" element={<AdminHome />} />
            <Route path="/admin/settings" element={<AdminSettings />} />
          </Route>

          {/* Tenant (GraphQL) */}
          <Route path="/t/:tenantId" element={<TenantLayout />}>
            <Route index element={<TenantHome />} />
            <Route path="orders" element={<TenantOrders />} />
          </Route>
        </Route>

        {/* Fallback */}
        <Route path="*" element={<Navigate to="/404" replace />} />
      </Routes>
    </React.Suspense>
  );
}
