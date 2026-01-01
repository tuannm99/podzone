import React from 'react';
import { Layout, Space, Button, Typography, Tag } from 'antd';
import { Link, useLocation, useParams } from 'react-router-dom';
import { logout } from '../services/auth';

const { Text } = Typography;

export default function TenantHeader() {
  const { tenantId } = useParams();
  const location = useLocation();

  return (
    <Layout.Header
      style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        padding: '0 16px',
        background: '#ffffff',
        borderBottom: '1px solid #f0f0f0',
      }}
    >
      <Space>
        <Link to={`/t/${tenantId}`} style={{ fontWeight: 700 }}>
          PODZONE Tenant
        </Link>
        <Tag color="blue">tenant: {tenantId}</Tag>
        <Text type="secondary">{location.pathname}</Text>
      </Space>

      <Button onClick={logout}>Logout</Button>
    </Layout.Header>
  );
}
