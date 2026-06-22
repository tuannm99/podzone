import {
  checkPlatformPermission,
  getPlatformUserPermissionBoundary,
  getPolicy,
  getRolePermissionBoundary,
  getRoleTrustPolicy,
  getTenantUserPermissionBoundary,
  listGroupInlinePolicies,
  listGroupMembers,
  listGroupPolicies,
  listGroups,
  listOrganizations,
  listPlatformUserInlinePolicies,
  listPlatformUserPolicies,
  listPolicies,
  listPolicyAttachments,
  listPolicyVersions,
  listServiceControlPolicies,
  listTenantUserInlinePolicies,
  listTenantUserPolicies,
  listUserTenants,
} from '@/services/iam';
import { prettyJSON } from './presentation';
import type { AdminIamState } from './useAdminIamState';

export function useAdminIamLoaders(state: AdminIamState, userID: number) {
  const {
    allowed,
    setAllowed,
    setLoading,
    setPageError,
    setMemberships,
    setOrganizations,
    setPolicies,
    selectedOrgId,
    setSelectedOrgId,
    selectedPolicyName,
    setSelectedPolicyName,
    selectedGroupId,
    groupScope,
    groupTenantId,
    setGroups,
    orgTenantId,
    setOrgTenantId,
    setGroupTenantId,
    simTenantId,
    setSimTenantId,
    principalTenantId,
    setPrincipalTenantId,
    shortcutTenantId,
    setShortcutTenantId,
    setPolicyDetail,
    setPolicyVersions,
    setPolicyAttachments,
    setOrgPolicies,
    setGroupMembers,
    setGroupPolicies,
    setGroupInlinePolicies,
    trustRoleName,
    setTrustJson,
    setRoleBoundary,
    setTrustBoundaryPolicyName,
    principalMode,
    principalPlatformUserId,
    principalTenantUserId,
    setPlatformUserPolicies,
    setPlatformUserInlinePolicies,
    setPlatformUserBoundary,
    setTenantUserPolicies,
    setTenantUserInlinePolicies,
    setTenantUserBoundary,
    setPrincipalBoundaryPolicyName,
  } = state;

  const loadBootstrap = async () => {
    setLoading(true);
    setPageError('');
    const permission = await checkPlatformPermission('platform:manage_roles');
    if (!permission.success) {
      setLoading(false);
      setPageError(permission.message);
      return;
    }
    setAllowed(permission.data);
    if (!permission.data) {
      setLoading(false);
      setPageError('Missing permission: platform:manage_roles');
      return;
    }

    const [tenantResult, orgResult, policyResult] = await Promise.all([
      listUserTenants(userID),
      listOrganizations(),
      listPolicies(),
    ]);

    if (!tenantResult.success) {
      setPageError(tenantResult.message);
    } else {
      setMemberships(tenantResult.data);
      const firstTenantID = tenantResult.data[0]?.tenantId;
      if (!orgTenantId() && firstTenantID) setOrgTenantId(firstTenantID);
      if (!groupTenantId() && firstTenantID) setGroupTenantId(firstTenantID);
      if (!simTenantId() && firstTenantID) setSimTenantId(firstTenantID);
      if (!principalTenantId() && firstTenantID) {
        setPrincipalTenantId(firstTenantID);
      }
      if (!shortcutTenantId() && firstTenantID) setShortcutTenantId(firstTenantID);
    }
    if (!orgResult.success) {
      setPageError(orgResult.message);
    } else {
      setOrganizations(orgResult.data);
      if (!selectedOrgId() && orgResult.data[0]?.id) {
        setSelectedOrgId(orgResult.data[0].id);
      }
    }
    if (!policyResult.success) {
      setPageError(policyResult.message);
    } else {
      setPolicies(policyResult.data);
      if (!selectedPolicyName() && policyResult.data[0]?.name) {
        setSelectedPolicyName(policyResult.data[0].name);
      }
    }
    setLoading(false);
  };

  const loadGroupsForScope = async () => {
    if (!allowed()) return;
    const result = await listGroups(
      groupScope(),
      groupScope() === 'tenant' ? groupTenantId().trim() : undefined
    );
    if (!result.success) {
      setPageError(result.message);
      return;
    }
    setGroups(result.data);
  };

  const loadSelectedPolicy = async () => {
    const name = selectedPolicyName().trim();
    if (!name) {
      setPolicyDetail(undefined);
      setPolicyVersions([]);
      setPolicyAttachments([]);
      return;
    }
    const [policyResult, versionsResult, attachmentsResult] =
      await Promise.all([
        getPolicy(name),
        listPolicyVersions(name),
        listPolicyAttachments(name),
      ]);
    if (policyResult.success) setPolicyDetail(policyResult.data);
    else setPageError(policyResult.message);
    if (versionsResult.success) setPolicyVersions(versionsResult.data);
    else setPageError(versionsResult.message);
    if (attachmentsResult.success) setPolicyAttachments(attachmentsResult.data);
    else setPageError(attachmentsResult.message);
  };

  const loadSelectedOrganization = async () => {
    const orgId = selectedOrgId().trim();
    if (!orgId) {
      setOrgPolicies([]);
      return;
    }
    const result = await listServiceControlPolicies(orgId);
    if (!result.success) {
      setPageError(result.message);
      return;
    }
    setOrgPolicies(result.data);
  };

  const loadSelectedGroup = async () => {
    const raw = selectedGroupId().trim();
    if (!raw) {
      setGroupMembers([]);
      setGroupPolicies([]);
      return;
    }
    const groupId = Number.parseInt(raw, 10);
    if (!Number.isFinite(groupId) || groupId <= 0) return;
    const [membersResult, policiesResult, inlineResult] = await Promise.all([
      listGroupMembers(groupId),
      listGroupPolicies(groupId),
      listGroupInlinePolicies(groupId),
    ]);
    if (membersResult.success) setGroupMembers(membersResult.data);
    else setPageError(membersResult.message);
    if (policiesResult.success) setGroupPolicies(policiesResult.data);
    else setPageError(policiesResult.message);
    if (inlineResult.success) setGroupInlinePolicies(inlineResult.data);
    else setPageError(inlineResult.message);
  };

  const loadTrustPolicy = async () => {
    const roleName = trustRoleName().trim();
    if (!roleName) return;
    const result = await getRoleTrustPolicy(roleName);
    if (!result.success) {
      setPageError(result.message);
      return;
    }
    setTrustJson(prettyJSON(result.data));
    const boundaryResult = await getRolePermissionBoundary(roleName);
    if (boundaryResult.success) {
      setRoleBoundary(boundaryResult.data);
      setTrustBoundaryPolicyName(boundaryResult.data?.policyName || '');
    } else {
      setPageError(boundaryResult.message);
    }
  };

  const loadPrincipalControls = async () => {
    if (!allowed()) return;
    if (principalMode() === 'platform') {
      const targetUserId = Number.parseInt(principalPlatformUserId().trim(), 10);
      if (!Number.isFinite(targetUserId) || targetUserId <= 0) {
        setPlatformUserPolicies([]);
        setPlatformUserInlinePolicies([]);
        setPlatformUserBoundary(null);
        return;
      }
      const [policiesResult, inlineResult, boundaryResult] = await Promise.all([
        listPlatformUserPolicies(targetUserId),
        listPlatformUserInlinePolicies(targetUserId),
        getPlatformUserPermissionBoundary(targetUserId),
      ]);
      if (policiesResult.success) setPlatformUserPolicies(policiesResult.data);
      else setPageError(policiesResult.message);
      if (inlineResult.success) setPlatformUserInlinePolicies(inlineResult.data);
      else setPageError(inlineResult.message);
      if (boundaryResult.success) {
        setPlatformUserBoundary(boundaryResult.data);
        setPrincipalBoundaryPolicyName(boundaryResult.data?.policyName || '');
      } else {
        setPageError(boundaryResult.message);
      }
      return;
    }

    const targetUserId = Number.parseInt(principalTenantUserId().trim(), 10);
    const tenantId = principalTenantId().trim();
    if (!tenantId || !Number.isFinite(targetUserId) || targetUserId <= 0) {
      setTenantUserPolicies([]);
      setTenantUserInlinePolicies([]);
      setTenantUserBoundary(null);
      return;
    }
    const [policiesResult, inlineResult, boundaryResult] = await Promise.all([
      listTenantUserPolicies(tenantId, targetUserId),
      listTenantUserInlinePolicies(tenantId, targetUserId),
      getTenantUserPermissionBoundary(tenantId, targetUserId),
    ]);
    if (policiesResult.success) setTenantUserPolicies(policiesResult.data);
    else setPageError(policiesResult.message);
    if (inlineResult.success) setTenantUserInlinePolicies(inlineResult.data);
    else setPageError(inlineResult.message);
    if (boundaryResult.success) {
      setTenantUserBoundary(boundaryResult.data);
      setPrincipalBoundaryPolicyName(boundaryResult.data?.policyName || '');
    } else {
      setPageError(boundaryResult.message);
    }
  };

  return {
    loadBootstrap,
    loadGroupsForScope,
    loadSelectedPolicy,
    loadSelectedOrganization,
    loadSelectedGroup,
    loadTrustPolicy,
    loadPrincipalControls,
  };
}

export type AdminIamLoaders = ReturnType<typeof useAdminIamLoaders>;
