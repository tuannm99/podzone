import React from 'react';
import { Card, Typography } from 'antd';
import { useParams } from 'react-router-dom';

const { Title, Paragraph } = Typography;

export default function TenantOrders() {
  const { tenantId } = useParams();

  return (
    <div>
      <Title level={2}>Orders</Title>
      <Card>
        <Paragraph>
          Orders page for tenant <b>{tenantId}</b>. Replace with real
          useQuery(...) later.
        </Paragraph>
      </Card>
    </div>
  );
}
