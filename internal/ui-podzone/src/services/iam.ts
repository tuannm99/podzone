import { http, type HttpError } from './http';

export type TenantMembership = {
  tenantId: string;
  userId: number;
  roleId?: number;
  roleName: string;
  status: string;
  createdAt?: string;
  updatedAt?: string;
};

export type IamResult<T> =
  | { success: true; data: T }
  | { success: false; message: string };

export type UpsertTenantMemberPayload = {
  tenantId: string;
  userId: number;
  roleName: string;
};

export type UpsertTenantMemberIdentityPayload = {
  tenantId: string;
  identity: string;
  roleName: string;
};

export type TenantInfo = {
  id: string;
  slug: string;
  name: string;
  createdAt?: string;
  updatedAt?: string;
};

export type CreateTenantPayload = {
  ownerUserId?: number;
  slug: string;
  name: string;
};

export type CreateTenantResult = {
  tenant?: TenantInfo;
  ownerMembership?: TenantMembership;
};

export type PlatformRoleMembership = {
  userId: number;
  roleId?: number;
  roleName: string;
  status: string;
  createdAt?: string;
  updatedAt?: string;
};

export type TenantInvite = {
  id: string;
  tenantId: string;
  email: string;
  roleId?: number;
  roleName: string;
  status: string;
  invitedByUserId?: number;
  acceptedByUserId?: number;
  createdAt?: string;
  updatedAt?: string;
  expiresAt?: string;
  acceptedAt?: string;
  revokedAt?: string;
};

export type CheckPermissionPayload = {
  tenantId: string;
  userId: number;
  permission: string;
};

export type UpsertPlatformRolePayload = {
  targetUserId: number;
  roleName: string;
};

export type CreateTenantInvitePayload = {
  tenantId: string;
  email: string;
  roleName: string;
};

function toFailure(error: unknown, fallback: string): IamResult<never> {
  const message =
    typeof error === 'object' &&
      error &&
      'message' in error &&
      typeof error.message === 'string'
      ? error.message
      : fallback;
  return { success: false, message };
}

export async function listUserTenants(
  userId: number
): Promise<IamResult<TenantMembership[]>> {
  try {
    const { data } = await http.get<{ memberships?: TenantMembership[] }>(
      `/auth/v1/iam/users/${userId}/tenants`
    );
    return { success: true, data: data.memberships || [] };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load user tenants');
  }
}

export async function createTenant(
  payload: CreateTenantPayload
): Promise<IamResult<CreateTenantResult>> {
  try {
    const { data } = await http.post<CreateTenantResult>(
      '/auth/v1/iam/tenants',
      payload
    );
    return { success: true, data };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to create tenant');
  }
}

export async function listPlatformRoles(
  targetUserId: number
): Promise<IamResult<PlatformRoleMembership[]>> {
  try {
    const { data } = await http.get<{ memberships?: PlatformRoleMembership[] }>(
      `/auth/v1/iam/platform-users/${targetUserId}/roles`
    );
    return { success: true, data: data.memberships || [] };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load platform roles');
  }
}

export async function listTenantMembers(
  tenantId: string
): Promise<IamResult<TenantMembership[]>> {
  try {
    const { data } = await http.get<{ memberships?: TenantMembership[] }>(
      `/auth/v1/iam/tenants/${tenantId}/members`
    );
    return { success: true, data: data.memberships || [] };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load tenant members');
  }
}

export async function checkPermission(
  payload: CheckPermissionPayload
): Promise<IamResult<boolean>> {
  try {
    const { data } = await http.post<{ allowed?: boolean }>(
      '/auth/v1/iam/permissions:check',
      payload
    );
    return { success: true, data: Boolean(data.allowed) };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to check permission');
  }
}

export async function checkPlatformPermission(
  permission: string
): Promise<IamResult<boolean>> {
  try {
    const { data } = await http.post<{ allowed?: boolean }>(
      '/auth/v1/iam/platform-permissions:check',
      { permission }
    );
    return { success: true, data: Boolean(data.allowed) };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to check platform permission');
  }
}

export async function upsertPlatformRole(
  payload: UpsertPlatformRolePayload
): Promise<IamResult<true>> {
  try {
    await http.post(
      `/auth/v1/iam/platform-users/${payload.targetUserId}/roles`,
      payload
    );
    return { success: true, data: true };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to save platform role');
  }
}

export async function removePlatformRole(
  targetUserId: number,
  roleName: string
): Promise<IamResult<true>> {
  try {
    await http.delete(
      `/auth/v1/iam/platform-users/${targetUserId}/roles/${roleName}`
    );
    return { success: true, data: true };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to remove platform role');
  }
}

export async function upsertTenantMember(
  payload: UpsertTenantMemberPayload
): Promise<IamResult<true>> {
  try {
    await http.post(
      `/auth/v1/iam/tenants/${payload.tenantId}/members`,
      payload
    );
    return { success: true, data: true };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to save tenant member');
  }
}

export async function upsertTenantMemberByIdentity(
  payload: UpsertTenantMemberIdentityPayload
): Promise<IamResult<{ userId: number; createdUser: boolean }>> {
  try {
    const { data } = await http.post<{
      userId?: number;
      createdUser?: boolean;
    }>(`/auth/v1/iam/tenants/${payload.tenantId}/members:resolve`, payload);
    return {
      success: true,
      data: {
        userId: Number(data.userId || 0),
        createdUser: Boolean(data.createdUser),
      },
    };
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to save tenant member by identity'
    );
  }
}

export async function removeTenantMember(
  tenantId: string,
  userId: number
): Promise<IamResult<true>> {
  try {
    await http.delete(`/auth/v1/iam/tenants/${tenantId}/members/${userId}`);
    return { success: true, data: true };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to remove tenant member');
  }
}

export async function createTenantInvite(
  payload: CreateTenantInvitePayload
): Promise<
  IamResult<{ invite?: TenantInvite; inviteToken: string; acceptUrl: string }>
> {
  try {
    const { data } = await http.post<{
      invite?: TenantInvite;
      inviteToken?: string;
      acceptUrl?: string;
    }>(`/auth/v1/iam/tenants/${payload.tenantId}/invites`, payload);
    return {
      success: true,
      data: {
        invite: data.invite,
        inviteToken: data.inviteToken || '',
        acceptUrl: data.acceptUrl || '',
      },
    };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to create tenant invite');
  }
}

export async function listTenantInvites(
  tenantId: string
): Promise<IamResult<TenantInvite[]>> {
  try {
    const { data } = await http.get<{ invites?: TenantInvite[] }>(
      `/auth/v1/iam/tenants/${tenantId}/invites`
    );
    return { success: true, data: data.invites || [] };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load tenant invites');
  }
}

export async function revokeTenantInvite(
  inviteId: string
): Promise<IamResult<true>> {
  try {
    await http.delete(`/auth/v1/iam/invites/${inviteId}`);
    return { success: true, data: true };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to revoke tenant invite');
  }
}

export async function acceptTenantInvite(
  inviteToken: string
): Promise<IamResult<TenantMembership>> {
  try {
    const { data } = await http.post<{ membership?: TenantMembership }>(
      '/auth/v1/iam/invites:accept',
      { inviteToken }
    );
    if (!data.membership) {
      return {
        success: false,
        message: 'Invite acceptance returned no membership',
      };
    }
    return { success: true, data: data.membership };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to accept tenant invite');
  }
}
