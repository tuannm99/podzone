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
  orgId?: string;
};

export type OrganizationInfo = {
  id: string;
  slug: string;
  name: string;
  createdAt?: string;
  updatedAt?: string;
};

export type PolicyCondition = {
  id?: number;
  operator: string;
  key: string;
  value: string;
};

export type PolicyStatement = {
  id?: number;
  policyId?: number;
  policyName?: string;
  effect: string;
  actionPattern: string;
  resourcePattern: string;
  conditions?: PolicyCondition[];
  createdAt?: string;
};

export type PolicyInfo = {
  id?: number;
  scope: string;
  name: string;
  description?: string;
  isSystem?: boolean;
  defaultVersion?: string;
  createdAt?: string;
  updatedAt?: string;
  statements?: PolicyStatement[];
};

export type PolicyVersionInfo = {
  id?: number;
  policyId?: number;
  policyName?: string;
  version: string;
  isDefault?: boolean;
  createdAt?: string;
};

export type RolePermissionBoundary = {
  roleId?: number;
  roleName: string;
  policyId?: number;
  policyName: string;
  createdAt?: string;
};

export type UserInlinePolicy = {
  scope: string;
  tenantId?: string;
  userId: number;
  name: string;
  description?: string;
  statements?: PolicyStatement[];
  createdAt?: string;
  updatedAt?: string;
};

export type GroupInlinePolicy = {
  groupId: number;
  name: string;
  description?: string;
  statements?: PolicyStatement[];
  createdAt?: string;
  updatedAt?: string;
};

export type PermissionBoundary = {
  scope: string;
  tenantId?: string;
  userId: number;
  policyId?: number;
  policyName: string;
  createdAt?: string;
};

export type PolicyAttachmentInfo = {
  attachmentType: string;
  scope?: string;
  tenantId?: string;
  roleId?: number;
  roleName?: string;
  userId?: number;
  groupId?: number;
  groupName?: string;
  createdAt?: string;
};

export type GroupInfo = {
  id?: number;
  scope: string;
  tenantId?: string;
  name: string;
  description?: string;
  isSystem?: boolean;
  createdAt?: string;
  updatedAt?: string;
};

export type RoleTrustStatement = {
  id?: number;
  roleId?: number;
  effect: string;
  principalType: string;
  principalPattern: string;
  tenantPattern?: string;
  externalIdPattern?: string;
  createdAt?: string;
};

export type SimulateMatchedStatement = {
  policyName: string;
  effect: string;
  actionPattern: string;
  resourcePattern: string;
  conditions?: PolicyCondition[];
  source: string;
};

export type SimulateDecisionLayer = {
  layer: string;
  allowed: boolean;
  reason: string;
  matchedStatements?: SimulateMatchedStatement[];
};

export type SimulateAccessResult = {
  allowed: boolean;
  decisionSource: string;
  reason: string;
  matchedStatements?: SimulateMatchedStatement[];
  layers?: SimulateDecisionLayer[];
};

export type CreateTenantPayload = {
  ownerUserId?: number;
  slug: string;
  name: string;
};

export type CreateOrganizationPayload = {
  name: string;
  slug: string;
};

export type CreatePolicyPayload = {
  scope: string;
  name: string;
  description?: string;
  statements: PolicyStatement[];
};

export type CreatePolicyVersionPayload = {
  name: string;
  statements: PolicyStatement[];
  setAsDefault?: boolean;
};

export type CreateGroupPayload = {
  scope: string;
  tenantId?: string;
  name: string;
  description?: string;
};

export type PutRoleTrustPolicyPayload = {
  roleName: string;
  statements: RoleTrustStatement[];
};

export type SimulateAccessPayload = {
  scope: string;
  tenantId?: string;
  userId: number;
  action: string;
  resource: string;
  useAssumedRole?: boolean;
  assumedRoleSession?: Record<string, unknown>;
  sessionPolicy?: PolicyStatement[];
  attributes?: Record<string, string>;
  servicePrincipal?: string;
  sessionTags?: Record<string, string>;
};

export type CreateTenantResult = {
  tenant?: TenantInfo;
  ownerMembership?: TenantMembership;
};

export type CreateOrganizationResult = {
  organization?: OrganizationInfo;
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

export async function createOrganization(
  payload: CreateOrganizationPayload
): Promise<IamResult<CreateOrganizationResult>> {
  try {
    const { data } = await http.post<CreateOrganizationResult>(
      '/auth/v1/iam/organizations',
      payload
    );
    return { success: true, data };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to create organization');
  }
}

export async function listOrganizations(): Promise<
  IamResult<OrganizationInfo[]>
> {
  try {
    const { data } = await http.get<{ organizations?: OrganizationInfo[] }>(
      '/auth/v1/iam/organizations'
    );
    return { success: true, data: data.organizations || [] };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load organizations');
  }
}

export async function attachTenantToOrganization(
  orgId: string,
  tenantId: string
): Promise<IamResult<true>> {
  try {
    await http.post(`/auth/v1/iam/organizations/${orgId}/tenants/${tenantId}`, {
      orgId,
      tenantId,
    });
    return { success: true, data: true };
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to attach tenant to organization'
    );
  }
}

export async function detachTenantFromOrganization(
  orgId: string,
  tenantId: string
): Promise<IamResult<true>> {
  try {
    await http.delete(
      `/auth/v1/iam/organizations/${orgId}/tenants/${tenantId}`
    );
    return { success: true, data: true };
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to detach tenant from organization'
    );
  }
}

export async function attachServiceControlPolicy(
  orgId: string,
  policyName: string
): Promise<IamResult<true>> {
  try {
    await http.post(
      `/auth/v1/iam/organizations/${orgId}/service-control-policies`,
      { orgId, policyName }
    );
    return { success: true, data: true };
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to attach service control policy'
    );
  }
}

export async function detachServiceControlPolicy(
  orgId: string,
  policyName: string
): Promise<IamResult<true>> {
  try {
    await http.delete(
      `/auth/v1/iam/organizations/${orgId}/service-control-policies/${policyName}`
    );
    return { success: true, data: true };
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to detach service control policy'
    );
  }
}

export async function listServiceControlPolicies(
  orgId: string
): Promise<IamResult<PolicyInfo[]>> {
  try {
    const { data } = await http.get<{ policies?: PolicyInfo[] }>(
      `/auth/v1/iam/organizations/${orgId}/service-control-policies`
    );
    return { success: true, data: data.policies || [] };
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to load service control policies'
    );
  }
}

export async function createPolicy(
  payload: CreatePolicyPayload
): Promise<IamResult<{ policy?: PolicyInfo; statements?: PolicyStatement[] }>> {
  try {
    const { data } = await http.post<{
      policy?: PolicyInfo;
      statements?: PolicyStatement[];
    }>('/auth/v1/iam/policies', payload);
    return { success: true, data };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to create policy');
  }
}

export async function listPolicies(): Promise<IamResult<PolicyInfo[]>> {
  try {
    const { data } = await http.get<{ policies?: PolicyInfo[] }>(
      '/auth/v1/iam/policies'
    );
    return { success: true, data: data.policies || [] };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load policies');
  }
}

export async function getPolicy(name: string): Promise<IamResult<PolicyInfo>> {
  try {
    const { data } = await http.get<{ policy?: PolicyInfo }>(
      `/auth/v1/iam/policies/${name}`
    );
    if (!data.policy) {
      return { success: false, message: 'Policy not found' };
    }
    return { success: true, data: data.policy };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load policy');
  }
}

export async function createPolicyVersion(
  payload: CreatePolicyVersionPayload
): Promise<
  IamResult<{
    policyVersion?: PolicyVersionInfo;
    statements?: PolicyStatement[];
  }>
> {
  try {
    const { data } = await http.post<{
      policyVersion?: PolicyVersionInfo;
      statements?: PolicyStatement[];
    }>(`/auth/v1/iam/policies/${payload.name}/versions`, payload);
    return { success: true, data };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to create policy version');
  }
}

export async function listPolicyVersions(
  name: string
): Promise<IamResult<PolicyVersionInfo[]>> {
  try {
    const { data } = await http.get<{ versions?: PolicyVersionInfo[] }>(
      `/auth/v1/iam/policies/${name}/versions`
    );
    return { success: true, data: data.versions || [] };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load policy versions');
  }
}

export async function setDefaultPolicyVersion(
  name: string,
  version: string
): Promise<IamResult<true>> {
  try {
    await http.post(
      `/auth/v1/iam/policies/${name}/versions/${version}:set-default`,
      {
        name,
        version,
      }
    );
    return { success: true, data: true };
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to set default policy version'
    );
  }
}

export async function deletePolicyVersion(
  name: string,
  version: string
): Promise<IamResult<true>> {
  try {
    await http.delete(`/auth/v1/iam/policies/${name}/versions/${version}`);
    return { success: true, data: true };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to delete policy version');
  }
}

export async function listPolicyAttachments(
  name: string
): Promise<IamResult<PolicyAttachmentInfo[]>> {
  try {
    const { data } = await http.get<{ attachments?: PolicyAttachmentInfo[] }>(
      `/auth/v1/iam/policies/${name}/attachments`
    );
    return { success: true, data: data.attachments || [] };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load policy attachments');
  }
}

export async function deletePolicy(name: string): Promise<IamResult<true>> {
  try {
    await http.delete(`/auth/v1/iam/policies/${name}`);
    return { success: true, data: true };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to delete policy');
  }
}

export async function createGroup(
  payload: CreateGroupPayload
): Promise<IamResult<{ group?: GroupInfo }>> {
  try {
    const { data } = await http.post<{ group?: GroupInfo }>(
      '/auth/v1/iam/groups',
      payload
    );
    return { success: true, data };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to create group');
  }
}

export async function listGroups(
  scope: string,
  tenantId?: string
): Promise<IamResult<GroupInfo[]>> {
  try {
    const params = new URLSearchParams({ scope });
    if (tenantId) params.set('tenantId', tenantId);
    const { data } = await http.get<{ groups?: GroupInfo[] }>(
      `/auth/v1/iam/groups?${params.toString()}`
    );
    return { success: true, data: data.groups || [] };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load groups');
  }
}

export async function deleteGroup(groupId: number): Promise<IamResult<true>> {
  try {
    await http.delete(`/auth/v1/iam/groups/${groupId}`);
    return { success: true, data: true };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to delete group');
  }
}

export async function addGroupMember(
  groupId: number,
  userId: number
): Promise<IamResult<true>> {
  try {
    await http.post(`/auth/v1/iam/groups/${groupId}/members`, {
      groupId,
      userId,
    });
    return { success: true, data: true };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to add group member');
  }
}

export async function listGroupMembers(
  groupId: number
): Promise<IamResult<number[]>> {
  try {
    const { data } = await http.get<{ userIds?: number[] }>(
      `/auth/v1/iam/groups/${groupId}/members`
    );
    return { success: true, data: data.userIds || [] };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load group members');
  }
}

export async function removeGroupMember(
  groupId: number,
  userId: number
): Promise<IamResult<true>> {
  try {
    await http.delete(`/auth/v1/iam/groups/${groupId}/members/${userId}`);
    return { success: true, data: true };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to remove group member');
  }
}

export async function attachGroupPolicy(
  groupId: number,
  policyName: string
): Promise<IamResult<true>> {
  try {
    await http.post(`/auth/v1/iam/groups/${groupId}/policies`, {
      groupId,
      policyName,
    });
    return { success: true, data: true };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to attach group policy');
  }
}

export async function listGroupPolicies(
  groupId: number
): Promise<IamResult<PolicyInfo[]>> {
  try {
    const { data } = await http.get<{ policies?: PolicyInfo[] }>(
      `/auth/v1/iam/groups/${groupId}/policies`
    );
    return { success: true, data: data.policies || [] };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load group policies');
  }
}

export async function detachGroupPolicy(
  groupId: number,
  policyName: string
): Promise<IamResult<true>> {
  try {
    await http.delete(`/auth/v1/iam/groups/${groupId}/policies/${policyName}`);
    return { success: true, data: true };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to detach group policy');
  }
}

export async function listGroupInlinePolicies(
  groupId: number
): Promise<IamResult<GroupInlinePolicy[]>> {
  try {
    const { data } = await http.get<{ policies?: GroupInlinePolicy[] }>(
      `/auth/v1/iam/groups/${groupId}/inline-policies`
    );
    return { success: true, data: data.policies || [] };
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to load group inline policies'
    );
  }
}

export async function putGroupInlinePolicy(
  groupId: number,
  name: string,
  description: string,
  statements: PolicyStatement[]
): Promise<IamResult<true>> {
  try {
    await http.put(`/auth/v1/iam/groups/${groupId}/inline-policies/${name}`, {
      groupId,
      name,
      description,
      statements,
    });
    return { success: true, data: true };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to save group inline policy');
  }
}

export async function deleteGroupInlinePolicy(
  groupId: number,
  name: string
): Promise<IamResult<true>> {
  try {
    await http.delete(`/auth/v1/iam/groups/${groupId}/inline-policies/${name}`);
    return { success: true, data: true };
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to delete group inline policy'
    );
  }
}

export async function listPlatformUserPolicies(
  targetUserId: number
): Promise<IamResult<PolicyInfo[]>> {
  try {
    const { data } = await http.get<{ policies?: PolicyInfo[] }>(
      `/auth/v1/iam/platform-users/${targetUserId}/policies`
    );
    return { success: true, data: data.policies || [] };
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to load platform user policies'
    );
  }
}

export async function attachPlatformUserPolicy(
  targetUserId: number,
  policyName: string
): Promise<IamResult<true>> {
  try {
    await http.post(`/auth/v1/iam/platform-users/${targetUserId}/policies`, {
      targetUserId,
      policyName,
    });
    return { success: true, data: true };
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to attach platform user policy'
    );
  }
}

export async function detachPlatformUserPolicy(
  targetUserId: number,
  policyName: string
): Promise<IamResult<true>> {
  try {
    await http.delete(
      `/auth/v1/iam/platform-users/${targetUserId}/policies/${policyName}`
    );
    return { success: true, data: true };
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to detach platform user policy'
    );
  }
}

export async function listPlatformUserInlinePolicies(
  targetUserId: number
): Promise<IamResult<UserInlinePolicy[]>> {
  try {
    const { data } = await http.get<{ policies?: UserInlinePolicy[] }>(
      `/auth/v1/iam/platform-users/${targetUserId}/inline-policies`
    );
    return { success: true, data: data.policies || [] };
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to load platform user inline policies'
    );
  }
}

export async function putPlatformUserInlinePolicy(
  targetUserId: number,
  name: string,
  description: string,
  statements: PolicyStatement[]
): Promise<IamResult<true>> {
  try {
    await http.put(
      `/auth/v1/iam/platform-users/${targetUserId}/inline-policies/${name}`,
      {
        targetUserId,
        name,
        description,
        statements,
      }
    );
    return { success: true, data: true };
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to save platform user inline policy'
    );
  }
}

export async function deletePlatformUserInlinePolicy(
  targetUserId: number,
  name: string
): Promise<IamResult<true>> {
  try {
    await http.delete(
      `/auth/v1/iam/platform-users/${targetUserId}/inline-policies/${name}`
    );
    return { success: true, data: true };
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to delete platform user inline policy'
    );
  }
}

export async function putPlatformUserPermissionBoundary(
  targetUserId: number,
  policyName: string
): Promise<IamResult<true>> {
  try {
    await http.put(
      `/auth/v1/iam/platform-users/${targetUserId}/permission-boundary`,
      {
        targetUserId,
        policyName,
      }
    );
    return { success: true, data: true };
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to save platform user boundary'
    );
  }
}

export async function getPlatformUserPermissionBoundary(
  targetUserId: number
): Promise<IamResult<PermissionBoundary | null>> {
  try {
    const { data } = await http.get<{ boundary?: PermissionBoundary }>(
      `/auth/v1/iam/platform-users/${targetUserId}/permission-boundary`
    );
    return { success: true, data: data.boundary || null };
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to load platform user boundary'
    );
  }
}

export async function deletePlatformUserPermissionBoundary(
  targetUserId: number
): Promise<IamResult<true>> {
  try {
    await http.delete(
      `/auth/v1/iam/platform-users/${targetUserId}/permission-boundary`
    );
    return { success: true, data: true };
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to delete platform user boundary'
    );
  }
}

export async function listTenantUserPolicies(
  tenantId: string,
  userId: number
): Promise<IamResult<PolicyInfo[]>> {
  try {
    const { data } = await http.get<{ policies?: PolicyInfo[] }>(
      `/auth/v1/iam/tenants/${tenantId}/members/${userId}/policies`
    );
    return { success: true, data: data.policies || [] };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load tenant user policies');
  }
}

export async function attachTenantUserPolicy(
  tenantId: string,
  userId: number,
  policyName: string
): Promise<IamResult<true>> {
  try {
    await http.post(
      `/auth/v1/iam/tenants/${tenantId}/members/${userId}/policies`,
      {
        tenantId,
        userId,
        policyName,
      }
    );
    return { success: true, data: true };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to attach tenant user policy');
  }
}

export async function detachTenantUserPolicy(
  tenantId: string,
  userId: number,
  policyName: string
): Promise<IamResult<true>> {
  try {
    await http.delete(
      `/auth/v1/iam/tenants/${tenantId}/members/${userId}/policies/${policyName}`
    );
    return { success: true, data: true };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to detach tenant user policy');
  }
}

export async function listTenantUserInlinePolicies(
  tenantId: string,
  userId: number
): Promise<IamResult<UserInlinePolicy[]>> {
  try {
    const { data } = await http.get<{ policies?: UserInlinePolicy[] }>(
      `/auth/v1/iam/tenants/${tenantId}/members/${userId}/inline-policies`
    );
    return { success: true, data: data.policies || [] };
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to load tenant user inline policies'
    );
  }
}

export async function putTenantUserInlinePolicy(
  tenantId: string,
  userId: number,
  name: string,
  description: string,
  statements: PolicyStatement[]
): Promise<IamResult<true>> {
  try {
    await http.put(
      `/auth/v1/iam/tenants/${tenantId}/members/${userId}/inline-policies/${name}`,
      {
        tenantId,
        userId,
        name,
        description,
        statements,
      }
    );
    return { success: true, data: true };
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to save tenant user inline policy'
    );
  }
}

export async function deleteTenantUserInlinePolicy(
  tenantId: string,
  userId: number,
  name: string
): Promise<IamResult<true>> {
  try {
    await http.delete(
      `/auth/v1/iam/tenants/${tenantId}/members/${userId}/inline-policies/${name}`
    );
    return { success: true, data: true };
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to delete tenant user inline policy'
    );
  }
}

export async function putTenantUserPermissionBoundary(
  tenantId: string,
  userId: number,
  policyName: string
): Promise<IamResult<true>> {
  try {
    await http.put(
      `/auth/v1/iam/tenants/${tenantId}/members/${userId}/permission-boundary`,
      {
        tenantId,
        userId,
        policyName,
      }
    );
    return { success: true, data: true };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to save tenant user boundary');
  }
}

export async function getTenantUserPermissionBoundary(
  tenantId: string,
  userId: number
): Promise<IamResult<PermissionBoundary | null>> {
  try {
    const { data } = await http.get<{ boundary?: PermissionBoundary }>(
      `/auth/v1/iam/tenants/${tenantId}/members/${userId}/permission-boundary`
    );
    return { success: true, data: data.boundary || null };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load tenant user boundary');
  }
}

export async function deleteTenantUserPermissionBoundary(
  tenantId: string,
  userId: number
): Promise<IamResult<true>> {
  try {
    await http.delete(
      `/auth/v1/iam/tenants/${tenantId}/members/${userId}/permission-boundary`
    );
    return { success: true, data: true };
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to delete tenant user boundary'
    );
  }
}

export async function putRoleTrustPolicy(
  payload: PutRoleTrustPolicyPayload
): Promise<IamResult<true>> {
  try {
    await http.put(
      `/auth/v1/iam/roles/${payload.roleName}/trust-policy`,
      payload
    );
    return { success: true, data: true };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to save role trust policy');
  }
}

export async function getRoleTrustPolicy(
  roleName: string
): Promise<IamResult<RoleTrustStatement[]>> {
  try {
    const { data } = await http.get<{ statements?: RoleTrustStatement[] }>(
      `/auth/v1/iam/roles/${roleName}/trust-policy`
    );
    return { success: true, data: data.statements || [] };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load role trust policy');
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
    });
    return { success: true, data: true };
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to save role permission boundary'
    );
  }
}

export async function getRolePermissionBoundary(
  roleName: string
): Promise<IamResult<RolePermissionBoundary | null>> {
  try {
    const { data } = await http.get<{ boundary?: RolePermissionBoundary }>(
      `/auth/v1/iam/roles/${roleName}/permission-boundary`
    );
    return { success: true, data: data.boundary || null };
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to load role permission boundary'
    );
  }
}

export async function deleteRolePermissionBoundary(
  roleName: string
): Promise<IamResult<true>> {
  try {
    await http.delete(`/auth/v1/iam/roles/${roleName}/permission-boundary`);
    return { success: true, data: true };
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to delete role permission boundary'
    );
  }
}

export async function simulateAccess(
  payload: SimulateAccessPayload
): Promise<IamResult<SimulateAccessResult>> {
  try {
    const { data } = await http.post<SimulateAccessResult>(
      '/auth/v1/iam/access:simulate',
      payload
    );
    return { success: true, data };
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to simulate access');
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
