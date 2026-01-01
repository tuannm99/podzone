import React, { createContext, useContext, useEffect, useMemo } from 'react';
import { useParams } from 'react-router-dom';
import { tenantStorage } from '../services/tenantStorage';

const TenantContext = createContext({ tenantId: '' });

export function TenantProvider({ children }) {
  const { tenantId = '' } = useParams();

  useEffect(() => {
    if (tenantId) tenantStorage.setTenantID(tenantId);
  }, [tenantId]);

  const value = useMemo(() => ({ tenantId }), [tenantId]);
  return (
    <TenantContext.Provider value={value}>{children}</TenantContext.Provider>
  );
}

export function UseTenant() {
  return useContext(TenantContext);
}
