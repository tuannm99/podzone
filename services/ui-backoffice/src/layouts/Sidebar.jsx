import React from 'react';
import { Layout, Menu } from 'antd';
import { HomeOutlined, SettingOutlined } from '@ant-design/icons';
import { Link, useLocation } from 'react-router-dom';

const { Sider } = Layout;

const Sidebar = () => {
    const location = useLocation();

    return (
        <Sider width={256} className="bg-base-200">
            <Menu mode="inline" selectedKeys={[location.pathname]} style={{ height: '100%', borderRight: 0 }}>
                <Menu.Item key="/" icon={<HomeOutlined />}>
                    <Link to="/">Home</Link>
                </Menu.Item>
                <Menu.Item key="/settings" icon={<SettingOutlined />}>
                    <Link to="/settings">Settings</Link>
                </Menu.Item>
                <Menu.Item key="/settings/404">
                    <Link to="/settings/404">Settings404</Link>
                </Menu.Item>
            </Menu>
        </Sider>
    );
};

export default Sidebar;
