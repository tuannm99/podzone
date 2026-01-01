import React from 'react';
import { Card, Typography } from 'antd';

const { Title, Paragraph } = Typography;

const SettingsPage = () => {
  return (
    <div>
      <Title level={2}>Settings</Title>
      <Card>
        <Paragraph>
          Settings page placeholder. Bạn có thể đặt cấu hình / form / tabs ở
          đây.
        </Paragraph>
      </Card>
    </div>
  );
};

export default SettingsPage;
