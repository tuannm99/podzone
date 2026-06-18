import { ONBOARDING_API_URL } from './baseurl';
import { http, type HttpError } from './http';

export type StoreRequestStatus =
  | 'requested'
  | 'pending_approval'
  | 'queued'
  | 'provisioning'
  | 'ready'
  | 'failed'
  | 'rejected'
  | 'suspended'
  | 'archived';

export type StoreRequest = {
  id: string;
  workspace_id: string;
  name: string;
  subdomain: string;
  requested_by: string;
  status: StoreRequestStatus;
  store_id?: string;
  last_error?: string;
  created_at: string;
  updated_at: string;
};

type Result<T> =
  | { success: true; data: T }
  | { success: false; message: string };

function errorMessage(error: unknown): string {
  const requestError = error as HttpError;
  return requestError?.message || 'Onboarding request failed';
}

function tenantHeaders(tenantId: string) {
  return { 'X-Tenant-ID': tenantId.trim() };
}

export async function listStoreRequests(
  tenantId: string
): Promise<Result<StoreRequest[]>> {
  try {
    const response = await http.get<StoreRequest[]>(
      `${ONBOARDING_API_URL}/onboarding/v1/requests`,
      { headers: tenantHeaders(tenantId) }
    );
    return { success: true, data: response.data || [] };
  } catch (error) {
    return { success: false, message: errorMessage(error) };
  }
}

export async function createStoreRequest(input: {
  tenantId: string;
  name: string;
  subdomain: string;
}): Promise<Result<StoreRequest>> {
  try {
    const { tenantId, ...payload } = input;
    const response = await http.post<StoreRequest>(
      `${ONBOARDING_API_URL}/onboarding/v1/requests`,
      payload,
      { headers: tenantHeaders(tenantId) }
    );
    return { success: true, data: response.data };
  } catch (error) {
    return { success: false, message: errorMessage(error) };
  }
}

export async function retryStoreRequest(input: {
  tenantId: string;
  requestId: string;
}): Promise<Result<void>> {
  try {
    await http.post(
      `${ONBOARDING_API_URL}/onboarding/v1/requests/${encodeURIComponent(input.requestId)}/retry`,
      undefined,
      { headers: tenantHeaders(input.tenantId) }
    );
    return { success: true, data: undefined };
  } catch (error) {
    return { success: false, message: errorMessage(error) };
  }
}
