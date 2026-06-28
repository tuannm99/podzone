import { http, type HttpError } from '../http'
import { toFailure } from './result'
import type {
  TenantMembership,
  IamResult,
  UpsertTenantMemberPayload,
  UpsertTenantMemberIdentityPayload,
  CreateTenantPayload,
  CreateTenantResult,
  TenantInvite,
  CreateTenantInvitePayload,
} from './types'

export async function listUserTenants(
  userId: number
): Promise<IamResult<TenantMembership[]>> {
  try {
    const { data } = await http.get<{ memberships?: TenantMembership[] }>(
      `/auth/v1/iam/users/${userId}/tenants`
    )
    return { success: true, data: data.memberships || [] }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load user tenants')
  }
}

export async function createTenant(
  payload: CreateTenantPayload
): Promise<IamResult<CreateTenantResult>> {
  try {
    const { data } = await http.post<CreateTenantResult>(
      '/auth/v1/iam/tenants',
      payload
    )
    return { success: true, data }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to create tenant')
  }
}

export async function listTenantMembers(
  tenantId: string
): Promise<IamResult<TenantMembership[]>> {
  try {
    const { data } = await http.get<{ memberships?: TenantMembership[] }>(
      `/auth/v1/iam/tenants/${tenantId}/members`
    )
    return { success: true, data: data.memberships || [] }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load tenant members')
  }
}

export async function upsertTenantMember(
  payload: UpsertTenantMemberPayload
): Promise<IamResult<true>> {
  try {
    await http.post(`/auth/v1/iam/tenants/${payload.tenantId}/members`, payload)
    return { success: true, data: true }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to save tenant member')
  }
}

export async function upsertTenantMemberByIdentity(
  payload: UpsertTenantMemberIdentityPayload
): Promise<IamResult<{ userId: number; createdUser: boolean }>> {
  try {
    const { data } = await http.post<{
      userId?: number
      createdUser?: boolean
    }>(`/auth/v1/iam/tenants/${payload.tenantId}/members:resolve`, payload)
    return {
      success: true,
      data: {
        userId: Number(data.userId || 0),
        createdUser: Boolean(data.createdUser),
      },
    }
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to save tenant member by identity'
    )
  }
}

export async function removeTenantMember(
  tenantId: string,
  userId: number
): Promise<IamResult<true>> {
  try {
    await http.delete(`/auth/v1/iam/tenants/${tenantId}/members/${userId}`)
    return { success: true, data: true }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to remove tenant member')
  }
}

export async function createTenantInvite(
  payload: CreateTenantInvitePayload
): Promise<
  IamResult<{ invite?: TenantInvite; inviteToken: string; acceptUrl: string }>
> {
  try {
    const { data } = await http.post<{
      invite?: TenantInvite
      inviteToken?: string
      acceptUrl?: string
    }>(`/auth/v1/iam/tenants/${payload.tenantId}/invites`, payload)
    return {
      success: true,
      data: {
        invite: data.invite,
        inviteToken: data.inviteToken || '',
        acceptUrl: data.acceptUrl || '',
      },
    }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to create tenant invite')
  }
}

export async function listTenantInvites(
  tenantId: string
): Promise<IamResult<TenantInvite[]>> {
  try {
    const { data } = await http.get<{ invites?: TenantInvite[] }>(
      `/auth/v1/iam/tenants/${tenantId}/invites`
    )
    return { success: true, data: data.invites || [] }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load tenant invites')
  }
}

export async function revokeTenantInvite(
  inviteId: string
): Promise<IamResult<true>> {
  try {
    await http.delete(`/auth/v1/iam/invites/${inviteId}`)
    return { success: true, data: true }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to revoke tenant invite')
  }
}

export async function acceptTenantInvite(
  inviteToken: string
): Promise<IamResult<TenantMembership>> {
  try {
    const { data } = await http.post<{ membership?: TenantMembership }>(
      '/auth/v1/iam/invites:accept',
      { inviteToken }
    )
    if (!data.membership) {
      return {
        success: false,
        message: 'Invite acceptance returned no membership',
      }
    }
    return { success: true, data: data.membership }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to accept tenant invite')
  }
}
