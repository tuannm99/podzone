import React from 'react';
import { Layout as AntLayout } from 'antd';
import Header from './Header';
import Sidebar from './Sidebar';

const { Content } = AntLayout;

const AppLayout = ({ children }) => {
    return (
        <AntLayout style={{ minHeight: '100vh' }}>
            <Header />
            <AntLayout>
                <Sidebar />
                <Content style={{ padding: 24, overflow: 'auto' }}>{children}</Content>
            </AntLayout>
        </AntLayout>
    );
};

export default AppLayout;
