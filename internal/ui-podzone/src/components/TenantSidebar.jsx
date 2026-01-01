import React from 'react';
import { Layout, Menu } from 'antd';
import { Link, useLocation, useParams } from 'react-router-dom';

export default function TenantSidebar() {
  const { tenantId } = useParams();
  const location = useLocation();
  const path = location.pathname;

  const base = `/t/${tenantId}`;
  const selectedKey = path.startsWith(`${base}/orders`)
    ? `${base}/orders`
    : base;

  return (
    <Layout.Sider
      width={240}
      style={{ background: '#fff', borderRight: '1px solid #f0f0f0' }}
    >
      <Menu
        mode="inline"
        selectedKeys={[selectedKey]}
        style={{ height: '100%' }}
      >
        <Menu.Item key={base}>
          <Link to={base}>Home</Link>
        </Menu.Item>
        <Menu.Item key={`${base}/orders`}>
          <Link to={`${base}/orders`}>Orders</Link>
        </Menu.Item>
      </Menu>
    </Layout.Sider>
  );
}
