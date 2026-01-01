import React from 'react';
import { Card, Typography } from 'antd';

const { Title, Paragraph } = Typography;

export default function AdminSettings() {
  return (
    <div>
      <Title level={2}>Admin Settings</Title>
      <Card>
        <Paragraph>Admin settings placeholder.</Paragraph>
      </Card>
    </div>
  );
}
