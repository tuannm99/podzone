import React, { useState } from 'react';
import { Button, Input, Space } from 'antd';
import { useNavigate } from 'react-router-dom';
import { tenantStorage } from '../services/tenantStorage';

export default function TenantSwitcher() {
  const [tenantId, setTenantId] = useState(() => tenantStorage.getTenantID());
  const navigate = useNavigate();

  return (
    <Space>
      <Input
        style={{ width: 220 }}
        placeholder="tenant id"
        value={tenantId}
        onChange={(e) => setTenantId(e.target.value)}
        allowClear
      />
      <Button
        type="primary"
        disabled={!tenantId}
        onClick={() => {
          tenantStorage.setTenantID(tenantId);
          navigate(`/t/${tenantId}`, { replace: false });
        }}
      >
        Go
      </Button>
    </Space>
  );
}
