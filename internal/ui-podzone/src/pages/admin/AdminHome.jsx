import React from 'react';
import { Card, Typography } from 'antd';

const { Title, Paragraph } = Typography;

export default function AdminHome() {
  return (
    <div>
      <Title level={2}>Admin Dashboard</Title>
      <Card>
        <Paragraph>
          Admin area (REST). Use TenantSwitcher in header to jump to a tenant
          route.
        </Paragraph>
      </Card>
    </div>
  );
}
