import { http, type HttpError } from '../http'
import { toFailure } from './result'
import type {
  IamResult,
  RolePermissionBoundary,
  RoleTrustStatement,
  PutRoleTrustPolicyPayload,
  PlatformRoleMembership,
  UpsertPlatformRolePayload,
} from './types'

export async function putRoleTrustPolicy(
  payload: PutRoleTrustPolicyPayload
): Promise<IamResult<true>> {
  try {
    await http.put(
      `/auth/v1/iam/roles/${payload.roleName}/trust-policy`,
      payload
    )
    return { success: true, data: true }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to save role trust policy')
  }
}

export async function getRoleTrustPolicy(
  roleName: string
): Promise<IamResult<RoleTrustStatement[]>> {
  try {
    const { data } = await http.get<{ statements?: RoleTrustStatement[] }>(
      `/auth/v1/iam/roles/${roleName}/trust-policy`
    )
    return { success: true, data: data.statements || [] }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load role trust policy')
  }
}

export async function putRolePermissionBoundary(
  roleName: string,
  policyName: string
): Promise<IamResult<true>> {
  try {
    await http.put(`/auth/v1/iam/roles/${roleName}/permission-boundary`, {
      roleName,
      policyName,
    })
    return { success: true, data: true }
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to save role permission boundary'
    )
  }
}

export async function getRolePermissionBoundary(
  roleName: string
): Promise<IamResult<RolePermissionBoundary | null>> {
  try {
    const { data } = await http.get<{ boundary?: RolePermissionBoundary }>(
      `/auth/v1/iam/roles/${roleName}/permission-boundary`
    )
    return { success: true, data: data.boundary || null }
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to load role permission boundary'
    )
  }
}

export async function deleteRolePermissionBoundary(
  roleName: string
): Promise<IamResult<true>> {
  try {
    await http.delete(`/auth/v1/iam/roles/${roleName}/permission-boundary`)
    return { success: true, data: true }
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to delete role permission boundary'
    )
  }
}

export async function listPlatformRoles(
  targetUserId: number
): Promise<IamResult<PlatformRoleMembership[]>> {
  try {
    const { data } = await http.get<{ memberships?: PlatformRoleMembership[] }>(
      `/auth/v1/iam/platform-users/${targetUserId}/roles`
    )
    return { success: true, data: data.memberships || [] }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load platform roles')
  }
}

export async function upsertPlatformRole(
  payload: UpsertPlatformRolePayload
): Promise<IamResult<true>> {
  try {
    await http.post(
      `/auth/v1/iam/platform-users/${payload.targetUserId}/roles`,
      payload
    )
    return { success: true, data: true }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to save platform role')
  }
}

export async function removePlatformRole(
  targetUserId: number,
  roleName: string
): Promise<IamResult<true>> {
  try {
    await http.delete(
      `/auth/v1/iam/platform-users/${targetUserId}/roles/${roleName}`
    )
    return { success: true, data: true }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to remove platform role')
  }
}
