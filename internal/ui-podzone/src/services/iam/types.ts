export type TenantMembership = {
    tenantId: string
    userId: number
    roleId?: number
    roleName: string
    status: string
    createdAt?: string
    updatedAt?: string
}

export type IamResult<T> = { success: true; data: T } | { success: false; message: string }

export type UpsertTenantMemberPayload = {
    tenantId: string
    userId: number
    roleName: string
}

export type UpsertTenantMemberIdentityPayload = {
    tenantId: string
    identity: string
    roleName: string
}

export type TenantInfo = {
    id: string
    slug: string
    name: string
    createdAt?: string
    updatedAt?: string
    orgId?: string
}

export type OrganizationInfo = {
    id: string
    slug: string
    name: string
    rootUserId?: string
    createdAt?: string
    updatedAt?: string
}

export type OrganizationMembership = {
    orgId: string
    userId: string
    roleId: string
    roleName: string
    status: string
    createdAt?: string
    updatedAt?: string
}

export type PolicyCondition = {
    id?: number
    operator: string
    key: string
    value: string
}

export type PolicyStatement = {
    id?: number
    policyId?: number
    policyName?: string
    effect: string
    actionPattern: string
    resourcePattern: string
    conditions?: PolicyCondition[]
    createdAt?: string
}

export type PolicyInfo = {
    id?: number
    scope: string
    orgId?: string
    name: string
    description?: string
    isSystem?: boolean
    defaultVersion?: string
    createdAt?: string
    updatedAt?: string
    statements?: PolicyStatement[]
}

export type PolicyVersionInfo = {
    id?: number
    policyId?: number
    policyName?: string
    version: string
    isDefault?: boolean
    createdAt?: string
}

export type RolePermissionBoundary = {
    roleId?: number
    roleName: string
    policyId?: number
    policyName: string
    createdAt?: string
}

export type UserInlinePolicy = {
    scope: string
    tenantId?: string
    userId: number
    name: string
    description?: string
    statements?: PolicyStatement[]
    createdAt?: string
    updatedAt?: string
}

export type GroupInlinePolicy = {
    groupId: number
    name: string
    description?: string
    statements?: PolicyStatement[]
    createdAt?: string
    updatedAt?: string
}

export type PermissionBoundary = {
    scope: string
    tenantId?: string
    userId: number
    policyId?: number
    policyName: string
    createdAt?: string
}

export type PolicyAttachmentInfo = {
    attachmentType: string
    scope?: string
    tenantId?: string
    roleId?: number
    roleName?: string
    userId?: number
    groupId?: number
    groupName?: string
    createdAt?: string
}

export type GroupInfo = {
    id?: number
    scope: string
    orgId?: string
    tenantId?: string
    name: string
    description?: string
    isSystem?: boolean
    createdAt?: string
    updatedAt?: string
}

export type RoleTrustStatement = {
    id?: number
    roleId?: number
    effect: string
    principalType: string
    principalPattern: string
    tenantPattern?: string
    externalIdPattern?: string
    createdAt?: string
}

export type SimulateMatchedStatement = {
    policyName: string
    effect: string
    actionPattern: string
    resourcePattern: string
    conditions?: PolicyCondition[]
    source: string
}

export type SimulateDecisionLayer = {
    layer: string
    allowed: boolean
    reason: string
    matchedStatements?: SimulateMatchedStatement[]
}

export type SimulateAccessResult = {
    allowed: boolean
    decisionSource: string
    reason: string
    matchedStatements?: SimulateMatchedStatement[]
    layers?: SimulateDecisionLayer[]
}

export type CreateTenantPayload = {
    ownerUserId?: number
    slug: string
    name: string
}

export type CreateOrganizationPayload = {
    name: string
    slug: string
}

export type CreatePolicyPayload = {
    scope: string
    orgId?: string
    name: string
    description?: string
    statements: PolicyStatement[]
}

export type CreatePolicyVersionPayload = {
    name: string
    scope: string
    orgId?: string
    statements: PolicyStatement[]
    setAsDefault?: boolean
}

export type CreateGroupPayload = {
    scope: string
    orgId?: string
    tenantId?: string
    name: string
    description?: string
}

export type PolicyLocator = {
    scope: string
    orgId?: string
    name: string
}

export type PutRoleTrustPolicyPayload = {
    roleName: string
    statements: RoleTrustStatement[]
}

export type SimulateAccessPayload = {
    scope: string
    tenantId?: string
    userId: number
    action: string
    resource: string
    useAssumedRole?: boolean
    assumedRoleSession?: Record<string, unknown>
    sessionPolicy?: PolicyStatement[]
    attributes?: Record<string, string>
    servicePrincipal?: string
    sessionTags?: Record<string, string>
}

export type CreateTenantResult = {
    tenant?: TenantInfo
    ownerMembership?: TenantMembership
}

export type CreateOrganizationResult = {
    organization?: OrganizationInfo
}

export type PlatformRoleMembership = {
    userId: number
    roleId?: number
    roleName: string
    status: string
    createdAt?: string
    updatedAt?: string
}

export type DirectoryUser = {
    id: number
    email: string
    username: string
    displayName?: string
}

export type PermissionInfo = {
    id: number
    name: string
    resource: string
    action: string
}

export type DirectoryScope = {
    scope: 'platform' | 'organization' | 'tenant'
    orgId?: string
    tenantId?: string
}

export type TenantInvite = {
    id: string
    tenantId: string
    email: string
    roleId?: number
    roleName: string
    status: string
    invitedByUserId?: number
    acceptedByUserId?: number
    createdAt?: string
    updatedAt?: string
    expiresAt?: string
    acceptedAt?: string
    revokedAt?: string
}

export type UpsertPlatformRolePayload = {
    targetUserId: number
    roleName: string
}

export type CreateTenantInvitePayload = {
    tenantId: string
    email: string
    roleName: string
}
