import React from 'react';
import { Spin } from 'antd';

export default function PageFallback() {
  return (
    <div
      style={{
        height: 'calc(100vh - 64px)', // minus header
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      <Spin size="large" />
    </div>
  );
}
