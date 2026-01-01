import React, { useMemo } from 'react';
import { Layout } from 'antd';
import { Outlet } from 'react-router-dom';
import { ApolloProvider } from '@apollo/client';

import TenantHeader from '../components/TenantHeader';
import TenantSidebar from '../components/TenantSidebar';

import { TenantProvider } from '../tenant/tenant.context';
import { createTenantApolloClient } from '../graphql/apolloClient';

export default function TenantLayout() {
  const client = useMemo(() => createTenantApolloClient(), []);

  return (
    <TenantProvider>
      <ApolloProvider client={client}>
        <Layout style={{ minHeight: '100vh' }}>
          <TenantHeader />
          <Layout>
            <TenantSidebar />
            <Layout.Content style={{ padding: 24, overflow: 'auto' }}>
              <Outlet />
            </Layout.Content>
          </Layout>
        </Layout>
      </ApolloProvider>
    </TenantProvider>
  );
}
