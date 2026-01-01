import React from 'react';
import { Card, Typography } from 'antd';
import { useParams } from 'react-router-dom';

const { Title, Paragraph } = Typography;

export default function TenantHome() {
  const { tenantId } = useParams();

  return (
    <div>
      <Title level={2}>Tenant Home</Title>
      <Card>
        <Paragraph>
          Tenant = <b>{tenantId}</b>
        </Paragraph>
        <Paragraph>
          This area is GraphQL-based. Requests will include X-Tenant-ID from
          route param.
        </Paragraph>
      </Card>
    </div>
  );
}
