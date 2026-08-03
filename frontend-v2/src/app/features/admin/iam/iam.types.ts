// Ported from frontend/packages/shared/services/iam/types.ts — the legacy
// IAMService REST surface's wire shape (camelCase JSON, per grpc-gateway's
// default protojson marshaling). Only the subset PZEP-0020's admin IAM
// section actually uses (Organizations, Platform Roles, Policies) — no
// tenant-directory type, no group/inline-policy/permission-boundary types,
// since those screens aren't in this PZEP's first cut.

export type OrganizationInfo = {
  id: string;
  slug: string;
  name: string;
  rootUserId?: string;
  createdAt?: string;
  updatedAt?: string;
};

export type OrganizationMembership = {
  orgId: string;
  userId: string;
  roleId: string;
  roleName: string;
  status: string;
  createdAt?: string;
  updatedAt?: string;
};

export type CreateOrganizationPayload = {
  name: string;
  slug: string;
};

export type PlatformRoleMembership = {
  userId: number;
  roleId?: number;
  roleName: string;
  status: string;
  createdAt?: string;
  updatedAt?: string;
};

export type DirectoryUser = {
  id: number;
  email: string;
  username: string;
  displayName?: string;
};

export type PolicyStatement = {
  id?: number;
  policyId?: number;
  policyName?: string;
  effect: string;
  actionPattern: string;
  resourcePattern: string;
  createdAt?: string;
};

export type PolicyInfo = {
  id?: number;
  scope: string;
  orgId?: string;
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
