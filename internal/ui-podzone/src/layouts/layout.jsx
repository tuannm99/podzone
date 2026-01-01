import React from 'react';
import { Layout } from 'antd';
import { Outlet } from 'react-router-dom';

import AppHeader from './Header';
import Sidebar from './Sidebar';

const AppLayout = () => {
  return (
    <Layout style={{ minHeight: '100vh' }}>
      <AppHeader />
      <Layout>
        <Sidebar />
        <Layout.Content style={{ padding: 24, overflow: 'auto' }}>
          <Outlet />
        </Layout.Content>
      </Layout>
    </Layout>
  );
};

export default AppLayout;
