import React from 'react';
import { Layout, Menu } from 'antd';
import { HomeOutlined, SettingOutlined } from '@ant-design/icons';
import { Link, useLocation } from 'react-router-dom';

const Sidebar = () => {
  const location = useLocation();
  const path = location.pathname;

  const selectedKey = path.startsWith('/settings') ? '/settings' : '/';

  return (
    <Layout.Sider
      width={256}
      style={{
        background: '#fff',
        borderRight: '1px solid #f0f0f0',
      }}
    >
      <Menu
        mode="inline"
        selectedKeys={[selectedKey]}
        style={{ height: '100%', borderRight: 0 }}
      >
        <Menu.Item key="/" icon={<HomeOutlined />}>
          <Link to="/">Home</Link>
        </Menu.Item>

        <Menu.Item key="/settings" icon={<SettingOutlined />}>
          <Link to="/settings">Settings</Link>
        </Menu.Item>
      </Menu>
    </Layout.Sider>
  );
};

export default Sidebar;
