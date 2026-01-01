import { ApolloClient, InMemoryCache, HttpLink, from } from '@apollo/client';
import { setContext } from '@apollo/client/link/context';
import { TENANT_GQL_URL } from '../services/baseurl';
import { tokenStorage } from '../services/tokenStorage';
import { tenantStorage } from '../services/tenantStorage';

export function createTenantApolloClient() {
  const httpLink = new HttpLink({ uri: TENANT_GQL_URL });

  const authLink = setContext((_, { headers }) => {
    const token = tokenStorage.getToken();
    const tenantId = tenantStorage.getTenantID();

    return {
      headers: {
        ...headers,
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...(tenantId ? { 'X-Tenant-ID': tenantId } : {}),
      },
    };
  });

  return new ApolloClient({
    link: from([authLink, httpLink]),
    cache: new InMemoryCache(),
  });
}
