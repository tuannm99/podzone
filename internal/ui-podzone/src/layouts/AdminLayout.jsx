import React from 'react';
import { Layout } from 'antd';
import { Outlet } from 'react-router-dom';

import AdminHeader from '../components/AdminHeader';
import AdminSidebar from '../components/AdminSidebar';

export default function AdminLayout() {
  return (
    <Layout style={{ minHeight: '100vh' }}>
      <AdminHeader />
      <Layout>
        <AdminSidebar />
        <Layout.Content style={{ padding: 24, overflow: 'auto' }}>
          <Outlet />
        </Layout.Content>
      </Layout>
    </Layout>
  );
}
