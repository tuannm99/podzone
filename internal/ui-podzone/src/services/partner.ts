import { http, type HttpError } from './http';

export type PartnerInfo = {
  id: string;
  tenantId: string;
  code: string;
  name: string;
  contactName: string;
  contactEmail: string;
  notes: string;
  partnerType: string;
  status: string;
  createdAt?: string;
  updatedAt?: string;
};

export type PartnerResult<T> =
  | { success: true; data: T }
  | { success: false; message: string };

export type CreatePartnerPayload = {
  tenantId: string;
  code?: string;
  name: string;
  contactName: string;
  contactEmail: string;
  notes: string;
  partnerType: string;
};

export type UpdatePartnerPayload = {
  id: string;
  name: string;
  contactName: string;
  contactEmail: string;
  notes: string;
  partnerType: string;
};

function toFailure(error: unknown, fallback: string): PartnerResult<never> {
  const message =
    typeof error === 'object' &&
    error &&
    'message' in error &&
    typeof error.message === 'string'
      ? error.message
      : fallback;
  return { success: false, message };
}

export async function listPartners(
  tenantId: string,
  partnerType = '',
  status = ''
): Promise<PartnerResult<PartnerInfo[]>> {
  try {
    const { data } = await http.get<{ partners?: PartnerInfo[] }>(
      '/partner/v1/partners',
      {
        params: {
          tenantId,
          partnerType: partnerType || undefined,
          status: status || undefined,
        },
      }
    );
    return { success: true, data: data.partners || [] };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load partners');
  }
}

export async function getPartner(
  id: string
): Promise<PartnerResult<PartnerInfo>> {
  try {
    const { data } = await http.get<PartnerInfo>(`/partner/v1/partners/${id}`);
    return { success: true, data };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load partner');
  }
}

export async function createPartner(
  payload: CreatePartnerPayload
): Promise<PartnerResult<PartnerInfo>> {
  try {
    const { data } = await http.post<PartnerInfo>('/partner/v1/partners', payload);
    return { success: true, data };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to create partner');
  }
}

export async function updatePartnerStatus(
  id: string,
  status: string
): Promise<PartnerResult<PartnerInfo>> {
  try {
    const { data } = await http.patch<PartnerInfo>(`/partner/v1/partners/${id}/status`, {
      status,
    });
    return { success: true, data };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to update partner status');
  }
}

export async function updatePartner(
  payload: UpdatePartnerPayload
): Promise<PartnerResult<PartnerInfo>> {
  try {
    const { data } = await http.put<PartnerInfo>(
      `/partner/v1/partners/${payload.id}`,
      payload
    );
    return { success: true, data };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to update partner');
  }
}
