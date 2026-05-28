import {
  For,
  Show,
  createEffect,
  createSignal,
  onMount,
} from 'solid-js';
import {
  addGroupMember,
  attachGroupPolicy,
  attachPlatformUserPolicy,
  attachServiceControlPolicy,
  attachTenantToOrganization,
  attachTenantUserPolicy,
  checkPlatformPermission,
  createGroup,
  createOrganization,
  createPolicy,
  createPolicyVersion,
  deleteGroup,
  deleteGroupInlinePolicy,
  deletePolicy,
  deletePolicyVersion,
  deletePlatformUserInlinePolicy,
  deletePlatformUserPermissionBoundary,
  deleteTenantUserInlinePolicy,
  deleteTenantUserPermissionBoundary,
  detachServiceControlPolicy,
  detachGroupPolicy,
  detachTenantFromOrganization,
  detachPlatformUserPolicy,
  detachTenantUserPolicy,
  getPolicy,
  getPlatformUserPermissionBoundary,
  getRolePermissionBoundary,
  getRoleTrustPolicy,
  getTenantUserPermissionBoundary,
  listGroupMembers,
  listGroupInlinePolicies,
  listGroupPolicies,
  listGroups,
  listOrganizations,
  listPolicies,
  listPolicyAttachments,
  listPolicyVersions,
  listPlatformUserInlinePolicies,
  listPlatformUserPolicies,
  listServiceControlPolicies,
  listTenantUserInlinePolicies,
  listTenantUserPolicies,
  listUserTenants,
  putGroupInlinePolicy,
  putPlatformUserInlinePolicy,
  putPlatformUserPermissionBoundary,
  putRolePermissionBoundary,
  putRoleTrustPolicy,
  putTenantUserInlinePolicy,
  putTenantUserPermissionBoundary,
  removeGroupMember,
  removePlatformRole,
  removeTenantMember,
  setDefaultPolicyVersion,
  simulateAccess,
  type GroupInfo,
  type GroupInlinePolicy,
  type OrganizationInfo,
  type PermissionBoundary,
  type PolicyAttachmentInfo,
  type PolicyInfo,
  type PolicyStatement,
  type PolicyVersionInfo,
  type RolePermissionBoundary,
  type RoleTrustStatement,
  type SimulateAccessResult,
  type TenantMembership,
  type UserInlinePolicy,
  deleteRolePermissionBoundary,
  upsertPlatformRole,
  upsertTenantMember,
} from '../../../services/iam';
import { tokenStorage } from '../../../services/tokenStorage';
import {
  EmptyBlock,
  ErrorAlert,
  InfoAlert,
  LoadingInline,
} from '../../components/common/Feedback';
import { PageShell } from '../../components/common/PageShell';
import { AdminIamPolicyProvider } from './admin-iam/policy-context';
import { PoliciesPanel } from './admin-iam/PoliciesPanel';
import { AdminIamGroupProvider } from './admin-iam/group-context';
import { GroupsPanel } from './admin-iam/GroupsPanel';
import { AdminIamPrincipalProvider } from './admin-iam/principal-context';
import { PrincipalPoliciesPanel } from './admin-iam/PrincipalPoliciesPanel';
import { AdminIamTrustSimProvider } from './admin-iam/trust-sim-context';
import { TrustSimulationPanel } from './admin-iam/TrustSimulationPanel';
import {
  Badge,
  Button,
  Card,
  InputField,
  SelectField,
} from '../../components/common/Primitives';
import { SectionLead } from '../../components/common/SectionLead';
import { SectionTitle } from '../../components/common/SectionTitle';

const policyScopeOptions = [
  { name: 'Platform', value: 'platform' },
  { name: 'Tenant', value: 'tenant' },
];

const groupScopeOptions = [
  { name: 'Platform', value: 'platform' },
  { name: 'Tenant', value: 'tenant' },
];

const tenantRoleOptions = [
  { name: 'Workspace owner', value: 'tenant_owner' },
  { name: 'Workspace admin', value: 'tenant_admin' },
  { name: 'Workspace operator', value: 'tenant_editor' },
  { name: 'Workspace viewer', value: 'tenant_viewer' },
];

const platformRoleOptions = [
  { name: 'Platform owner', value: 'platform_owner' },
  { name: 'Platform admin', value: 'platform_admin' },
];

function prettyJSON(value: unknown) {
  return JSON.stringify(value, null, 2);
}

function parseJSONArray<T>(raw: string, label: string) {
  const parsed = JSON.parse(raw || '[]');
  if (!Array.isArray(parsed)) {
    throw new Error(`${label} must be a JSON array`);
  }
  return parsed as T[];
}

function parseJSONObject(raw: string, label: string) {
  const parsed = JSON.parse(raw || '{}');
  if (!parsed || Array.isArray(parsed) || typeof parsed !== 'object') {
    throw new Error(`${label} must be a JSON object`);
  }
  return parsed as Record<string, string>;
}

function attachmentColor(type: string) {
  if (type.includes('boundary')) return 'pink';
  if (type.includes('service_control')) return 'yellow';
  if (type.includes('group')) return 'green';
  return 'blue';
}

function simulationSourceColor(source: string) {
  const normalized = source.toLowerCase();
  if (normalized.includes('deny')) return 'red';
  if (normalized.includes('boundary')) return 'pink';
  if (normalized.includes('scp')) return 'yellow';
  if (normalized.includes('session')) return 'indigo';
  if (normalized.includes('group')) return 'green';
  return 'blue';
}

function simulationLayerTone(allowed: boolean, reason: string) {
  if (!allowed) {
    if (reason.toLowerCase().includes('deny')) {
      return 'border-red-200 bg-red-50';
    }
    return 'border-amber-200 bg-amber-50';
  }
  return 'border-green-200 bg-green-50';
}

function statementSourceLabel(source: string) {
  return source.replaceAll('_', ' ');
}

export default function AdminIamPage() {
  const userID = tokenStorage.getUserID() || 0;

  const [pageError, setPageError] = createSignal('');
  const [pageMessage, setPageMessage] = createSignal('');
  const [loading, setLoading] = createSignal(false);
  const [allowed, setAllowed] = createSignal(false);

  const [memberships, setMemberships] = createSignal<TenantMembership[]>([]);
  const [organizations, setOrganizations] = createSignal<OrganizationInfo[]>([]);
  const [policies, setPolicies] = createSignal<PolicyInfo[]>([]);
  const [groups, setGroups] = createSignal<GroupInfo[]>([]);

  const [selectedOrgId, setSelectedOrgId] = createSignal('');
  const [selectedPolicyName, setSelectedPolicyName] = createSignal('');
  const [selectedGroupId, setSelectedGroupId] = createSignal('');

  const [policyDetail, setPolicyDetail] = createSignal<PolicyInfo>();
  const [policyVersions, setPolicyVersions] = createSignal<PolicyVersionInfo[]>([]);
  const [policyAttachments, setPolicyAttachments] = createSignal<PolicyAttachmentInfo[]>([]);
  const [orgPolicies, setOrgPolicies] = createSignal<PolicyInfo[]>([]);
  const [groupMembers, setGroupMembers] = createSignal<number[]>([]);
  const [groupPolicies, setGroupPolicies] = createSignal<PolicyInfo[]>([]);
  const [groupInlinePolicies, setGroupInlinePolicies] = createSignal<GroupInlinePolicy[]>([]);
  const [roleBoundary, setRoleBoundary] = createSignal<RolePermissionBoundary | null>(null);
  const [platformUserPolicies, setPlatformUserPolicies] = createSignal<PolicyInfo[]>([]);
  const [tenantUserPolicies, setTenantUserPolicies] = createSignal<PolicyInfo[]>([]);
  const [platformUserInlinePolicies, setPlatformUserInlinePolicies] = createSignal<UserInlinePolicy[]>([]);
  const [tenantUserInlinePolicies, setTenantUserInlinePolicies] = createSignal<UserInlinePolicy[]>([]);
  const [platformUserBoundary, setPlatformUserBoundary] = createSignal<PermissionBoundary | null>(null);
  const [tenantUserBoundary, setTenantUserBoundary] = createSignal<PermissionBoundary | null>(null);
  const [simulation, setSimulation] = createSignal<SimulateAccessResult>();

  const [orgName, setOrgName] = createSignal('');
  const [orgSlug, setOrgSlug] = createSignal('');
  const [orgTenantId, setOrgTenantId] = createSignal('');
  const [orgPolicyName, setOrgPolicyName] = createSignal('');

  const [policyScope, setPolicyScope] = createSignal('platform');
  const [policyName, setPolicyName] = createSignal('');
  const [policyDescription, setPolicyDescription] = createSignal('');
  const [policyStatementsJson, setPolicyStatementsJson] = createSignal(
    prettyJSON([
      {
        effect: 'allow',
        actionPattern: 'order:read',
        resourcePattern: '*',
        conditions: [],
      },
    ])
  );
  const [policyVersionJson, setPolicyVersionJson] = createSignal(
    prettyJSON([
      {
        effect: 'allow',
        actionPattern: 'order:update',
        resourcePattern: '*',
        conditions: [],
      },
    ])
  );

  const [groupScope, setGroupScope] = createSignal('platform');
  const [groupTenantId, setGroupTenantId] = createSignal('');
  const [groupName, setGroupName] = createSignal('');
  const [groupDescription, setGroupDescription] = createSignal('');
  const [groupMemberUserId, setGroupMemberUserId] = createSignal('');
  const [groupPolicyName, setGroupPolicyName] = createSignal('');
  const [groupInlinePolicyName, setGroupInlinePolicyName] = createSignal('');
  const [groupInlinePolicyDescription, setGroupInlinePolicyDescription] = createSignal('');
  const [groupInlinePolicyJson, setGroupInlinePolicyJson] = createSignal(
    prettyJSON([
      {
        effect: 'allow',
        actionPattern: 'order:read',
        resourcePattern: '*',
        conditions: [],
      },
    ])
  );

  const [shortcutPlatformUserId, setShortcutPlatformUserId] = createSignal(userID ? String(userID) : '');
  const [shortcutPlatformRoleName, setShortcutPlatformRoleName] = createSignal(platformRoleOptions[1].value);
  const [shortcutTenantId, setShortcutTenantId] = createSignal('');
  const [shortcutTenantUserId, setShortcutTenantUserId] = createSignal('');
  const [shortcutTenantRoleName, setShortcutTenantRoleName] = createSignal(tenantRoleOptions[1].value);

  const [trustRoleName, setTrustRoleName] = createSignal('tenant_admin');
  const [trustBoundaryPolicyName, setTrustBoundaryPolicyName] = createSignal('');
  const [trustJson, setTrustJson] = createSignal(
    prettyJSON([
      {
        effect: 'allow',
        principalType: 'service',
        principalPattern: 'backoffice.podzone.internal',
        tenantPattern: '*',
        externalIdPattern: '',
      },
    ])
  );

  const [simScope, setSimScope] = createSignal('tenant');
  const [simTenantId, setSimTenantId] = createSignal('');
  const [simTargetUserId, setSimTargetUserId] = createSignal(userID ? String(userID) : '');
  const [simAction, setSimAction] = createSignal('order:update');
  const [simResource, setSimResource] = createSignal('*');
  const [simServicePrincipal, setSimServicePrincipal] = createSignal('');
  const [simAttributesJson, setSimAttributesJson] = createSignal(prettyJSON({}));
  const [simSessionTagsJson, setSimSessionTagsJson] = createSignal(
    prettyJSON({ team: 'ops', lane: 'priority' })
  );
  const [simSessionPolicyJson, setSimSessionPolicyJson] = createSignal(
    prettyJSON([
      {
        effect: 'allow',
        actionPattern: 'order:update',
        resourcePattern: '*',
        conditions: [],
      },
    ])
  );
  const [simAssumedRoleId, setSimAssumedRoleId] = createSignal('');
  const [simAssumedRoleScope, setSimAssumedRoleScope] = createSignal('tenant');
  const [simAssumedRoleName, setSimAssumedRoleName] = createSignal('');
  const [simAssumedRoleTenantId, setSimAssumedRoleTenantId] = createSignal('');
  const [simAssumedRoleSessionName, setSimAssumedRoleSessionName] = createSignal('');
  const [simAssumedRoleSourceIdentity, setSimAssumedRoleSourceIdentity] = createSignal('');
  const [simAssumedRoleServicePrincipal, setSimAssumedRoleServicePrincipal] = createSignal('');
  const [simAssumedRoleExpiresAt, setSimAssumedRoleExpiresAt] = createSignal('');

  const [principalMode, setPrincipalMode] = createSignal<'platform' | 'tenant'>('platform');
  const [principalPlatformUserId, setPrincipalPlatformUserId] = createSignal(userID ? String(userID) : '');
  const [principalTenantId, setPrincipalTenantId] = createSignal('');
  const [principalTenantUserId, setPrincipalTenantUserId] = createSignal('');
  const [principalManagedPolicyName, setPrincipalManagedPolicyName] = createSignal('');
  const [principalBoundaryPolicyName, setPrincipalBoundaryPolicyName] = createSignal('');
  const [principalInlinePolicyName, setPrincipalInlinePolicyName] = createSignal('');
  const [principalInlinePolicyDescription, setPrincipalInlinePolicyDescription] = createSignal('');
  const [principalInlinePolicyJson, setPrincipalInlinePolicyJson] = createSignal(
    prettyJSON([
      {
        effect: 'allow',
        actionPattern: 'order:read',
        resourcePattern: '*',
        conditions: [],
      },
    ])
  );

  const tenantOptions = () =>
    memberships().map((membership) => ({
      name: `${membership.tenantId} · ${membership.roleName}`,
      value: membership.tenantId,
    }));

  const policyOptions = () =>
    policies().map((item) => ({ name: `${item.name} · ${item.scope}`, value: item.name }));

  const organizationOptions = () =>
    organizations().map((item) => ({ name: `${item.slug} · ${item.name}`, value: item.id }));

  const groupOptions = () =>
    groups().map((item) => ({
      name: `${item.name}${item.tenantId ? ` · ${item.tenantId}` : ''}`,
      value: String(item.id || ''),
    }));

  const buildAssumedRoleSession = () => {
    const roleId = Number.parseInt(simAssumedRoleId().trim(), 10);
    if (!Number.isFinite(roleId) || roleId <= 0) return undefined;
    return {
      assumedRoleId: roleId,
      assumedRoleScope: simAssumedRoleScope().trim(),
      assumedRoleName: simAssumedRoleName().trim(),
      assumedRoleTenantId: simAssumedRoleTenantId().trim() || undefined,
      assumedRoleSessionName: simAssumedRoleSessionName().trim() || undefined,
      assumedRoleSourceIdentity:
        simAssumedRoleSourceIdentity().trim() || undefined,
      assumedRoleServicePrincipal:
        simAssumedRoleServicePrincipal().trim() || undefined,
      assumedRoleExpiresAt: simAssumedRoleExpiresAt().trim() || undefined,
      sessionTags: parseJSONObject(simSessionTagsJson(), 'Session tags'),
    };
  };

  createEffect(() => {
    if (
      simAssumedRoleScope() === 'tenant' &&
      simTenantId().trim() &&
      !simAssumedRoleTenantId().trim()
    ) {
      setSimAssumedRoleTenantId(simTenantId().trim());
    }
  });

  const applyServiceAssumePreset = () => {
    setSimScope('platform');
    setSimAction('order:update');
    setSimResource('*');
    setSimServicePrincipal('backoffice.podzone.internal');
    setSimAssumedRoleScope('platform');
    setSimAssumedRoleName('platform_admin');
    setSimAssumedRoleSourceIdentity('backoffice-admin');
    setSimAssumedRoleSessionName('service-assume');
    setSimAssumedRoleServicePrincipal('backoffice.podzone.internal');
    setSimAttributesJson(prettyJSON({ lane: 'priority' }));
    setSimSessionTagsJson(prettyJSON({ team: 'ops', path: 'service-assume' }));
  };

  const applyTenantAssumePreset = () => {
    setSimScope('tenant');
    setSimAction('order:update');
    setSimResource('*');
    setSimAssumedRoleScope('tenant');
    setSimAssumedRoleName('tenant_admin');
    setSimAssumedRoleTenantId(simTenantId().trim());
    setSimAssumedRoleSessionName('tenant-admin-review');
    setSimAssumedRoleSourceIdentity('store-ops');
    setSimAssumedRoleServicePrincipal('');
    setSimAttributesJson(prettyJSON({ lane: 'priority', region: 'us' }));
    setSimSessionTagsJson(prettyJSON({ team: 'ops', store: simTenantId().trim() || 'tenant' }));
  };

  const applyScopeDownDenyPreset = () => {
    setSimAction('order:update');
    setSimResource('*');
    setSimSessionPolicyJson(
      prettyJSON([
        {
          effect: 'deny',
          actionPattern: 'order:update',
          resourcePattern: '*',
          conditions: [],
        },
      ])
    );
    setSimAttributesJson(prettyJSON({ lane: 'restricted' }));
    setSimSessionTagsJson(prettyJSON({ team: 'ops', mode: 'scope-down' }));
  };

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
      setPageError('You do not have permission to manage IAM.');
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
      if (!orgTenantId() && tenantResult.data[0]?.tenantId) setOrgTenantId(tenantResult.data[0].tenantId);
      if (!groupTenantId() && tenantResult.data[0]?.tenantId) setGroupTenantId(tenantResult.data[0].tenantId);
      if (!simTenantId() && tenantResult.data[0]?.tenantId) setSimTenantId(tenantResult.data[0].tenantId);
      if (!principalTenantId() && tenantResult.data[0]?.tenantId) setPrincipalTenantId(tenantResult.data[0].tenantId);
      if (!shortcutTenantId() && tenantResult.data[0]?.tenantId) setShortcutTenantId(tenantResult.data[0].tenantId);
    }
    if (!orgResult.success) {
      setPageError(orgResult.message);
    } else {
      setOrganizations(orgResult.data);
      if (!selectedOrgId() && orgResult.data[0]?.id) setSelectedOrgId(orgResult.data[0].id);
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
    const result = await listGroups(groupScope(), groupScope() === 'tenant' ? groupTenantId().trim() : undefined);
    if (!result.success) {
      setPageError(result.message);
      return;
    }
    setGroups(result.data);
    if (!selectedGroupId() && result.data[0]?.id) {
      setSelectedGroupId(String(result.data[0].id));
    }
  };

  const loadSelectedPolicy = async () => {
    const name = selectedPolicyName().trim();
    if (!name) {
      setPolicyDetail(undefined);
      setPolicyVersions([]);
      setPolicyAttachments([]);
      return;
    }
    const [policyResult, versionsResult, attachmentsResult] = await Promise.all([
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
    const [membersResult, policiesResult] = await Promise.all([
      listGroupMembers(groupId),
      listGroupPolicies(groupId),
    ]);
    if (membersResult.success) setGroupMembers(membersResult.data);
    else setPageError(membersResult.message);
    if (policiesResult.success) setGroupPolicies(policiesResult.data);
    else setPageError(policiesResult.message);
    const inlineResult = await listGroupInlinePolicies(groupId);
    if (inlineResult.success) setGroupInlinePolicies(inlineResult.data);
    else setPageError(inlineResult.message);
  };

  const loadTrustPolicy = async () => {
    const roleName = trustRoleName().trim();
    if (!roleName) {
      return;
    }
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

  createEffect(() => {
    void selectedPolicyName();
    if (allowed()) void loadSelectedPolicy();
  });

  createEffect(() => {
    void selectedOrgId();
    if (allowed()) void loadSelectedOrganization();
  });

  createEffect(() => {
    void selectedGroupId();
    if (allowed()) void loadSelectedGroup();
  });

  createEffect(() => {
    void groupScope();
    void groupTenantId();
    if (allowed()) void loadGroupsForScope();
  });

  createEffect(() => {
    void principalMode();
    void principalPlatformUserId();
    void principalTenantId();
    void principalTenantUserId();
    if (allowed()) void loadPrincipalControls();
  });

  onMount(() => {
    void loadBootstrap();
  });

  const runAction = async (work: () => Promise<void>) => {
    setPageError('');
    setPageMessage('');
    try {
      await work();
    } catch (error) {
      setPageError(error instanceof Error ? error.message : 'Action failed');
    }
  };

  const submitCreateOrganization = async (event: SubmitEvent) => {
    event.preventDefault();
    await runAction(async () => {
      const result = await createOrganization({
        name: orgName().trim(),
        slug: orgSlug().trim(),
      });
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Created organization ${result.data.organization?.slug || orgName()}.`);
      setOrgName('');
      setOrgSlug('');
      await loadBootstrap();
    });
  };

  const handleAttachTenantToOrg = async () => {
    await runAction(async () => {
      const result = await attachTenantToOrganization(selectedOrgId().trim(), orgTenantId().trim());
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Attached tenant ${orgTenantId().trim()} to organization.`);
      await loadBootstrap();
    });
  };

  const handleDetachTenantFromOrg = async (tenantId: string) => {
    await runAction(async () => {
      const result = await detachTenantFromOrganization(selectedOrgId().trim(), tenantId);
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Detached tenant ${tenantId} from organization.`);
      await loadBootstrap();
    });
  };

  const handleAttachScp = async () => {
    await runAction(async () => {
      const result = await attachServiceControlPolicy(selectedOrgId().trim(), orgPolicyName().trim());
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Attached SCP ${orgPolicyName().trim()}.`);
      setOrgPolicyName('');
      await loadSelectedOrganization();
      await loadSelectedPolicy();
    });
  };

  const handleDetachScp = async (policyName: string) => {
    await runAction(async () => {
      const result = await detachServiceControlPolicy(selectedOrgId().trim(), policyName);
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Detached SCP ${policyName}.`);
      await loadSelectedOrganization();
      await loadSelectedPolicy();
    });
  };

  const submitCreatePolicy = async (event: SubmitEvent) => {
    event.preventDefault();
    await runAction(async () => {
      const statements = parseJSONArray<PolicyStatement>(policyStatementsJson(), 'Policy statements');
      const result = await createPolicy({
        scope: policyScope(),
        name: policyName().trim(),
        description: policyDescription().trim(),
        statements,
      });
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Created policy ${policyName().trim()}.`);
      setPolicyName('');
      setPolicyDescription('');
      await loadBootstrap();
    });
  };

  const handleCreatePolicyVersion = async () => {
    await runAction(async () => {
      const statements = parseJSONArray<PolicyStatement>(policyVersionJson(), 'Policy version statements');
      const result = await createPolicyVersion({
        name: selectedPolicyName().trim(),
        statements,
        setAsDefault: false,
      });
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Created a new version for ${selectedPolicyName().trim()}.`);
      await loadSelectedPolicy();
    });
  };

  const handleDeletePolicy = async () => {
    await runAction(async () => {
      const name = selectedPolicyName().trim();
      const result = await deletePolicy(name);
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Deleted policy ${name}.`);
      setSelectedPolicyName('');
      setPolicyDetail(undefined);
      setPolicyVersions([]);
      setPolicyAttachments([]);
      await loadBootstrap();
    });
  };

  const handleSetDefaultVersion = async (version: string) => {
    await runAction(async () => {
      const result = await setDefaultPolicyVersion(selectedPolicyName().trim(), version);
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Set ${version} as default for ${selectedPolicyName().trim()}.`);
      await loadSelectedPolicy();
    });
  };

  const handleDeleteVersion = async (version: string) => {
    await runAction(async () => {
      const result = await deletePolicyVersion(selectedPolicyName().trim(), version);
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Deleted policy version ${version}.`);
      await loadSelectedPolicy();
    });
  };

  const submitCreateGroup = async (event: SubmitEvent) => {
    event.preventDefault();
    await runAction(async () => {
      const result = await createGroup({
        scope: groupScope(),
        tenantId: groupScope() === 'tenant' ? groupTenantId().trim() : undefined,
        name: groupName().trim(),
        description: groupDescription().trim(),
      });
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Created group ${groupName().trim()}.`);
      setGroupName('');
      setGroupDescription('');
      await loadGroupsForScope();
    });
  };

  const handleAddGroupMember = async () => {
    await runAction(async () => {
      const groupId = Number.parseInt(selectedGroupId().trim(), 10);
      const userId = Number.parseInt(groupMemberUserId().trim(), 10);
      const result = await addGroupMember(groupId, userId);
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Added user ${userId} to group.`);
      setGroupMemberUserId('');
      await loadSelectedGroup();
    });
  };

  const handleRemoveGroupMember = async (userId: number) => {
    await runAction(async () => {
      const groupId = Number.parseInt(selectedGroupId().trim(), 10);
      const result = await removeGroupMember(groupId, userId);
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Removed user ${userId} from group.`);
      await loadSelectedGroup();
    });
  };

  const handleAttachGroupPolicy = async () => {
    await runAction(async () => {
      const groupId = Number.parseInt(selectedGroupId().trim(), 10);
      const result = await attachGroupPolicy(groupId, groupPolicyName().trim());
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Attached policy ${groupPolicyName().trim()} to group.`);
      setGroupPolicyName('');
      await loadSelectedGroup();
      await loadSelectedPolicy();
    });
  };

  const handleDetachGroupPolicy = async (policyName: string) => {
    await runAction(async () => {
      const groupId = Number.parseInt(selectedGroupId().trim(), 10);
      const result = await detachGroupPolicy(groupId, policyName);
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Detached policy ${policyName} from group.`);
      await loadSelectedGroup();
      await loadSelectedPolicy();
    });
  };

  const handleDeleteGroup = async () => {
    await runAction(async () => {
      const groupId = Number.parseInt(selectedGroupId().trim(), 10);
      const result = await deleteGroup(groupId);
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Deleted group ${selectedGroupId().trim()}.`);
      setSelectedGroupId('');
      setGroupMembers([]);
      setGroupPolicies([]);
      setGroupInlinePolicies([]);
      await loadGroupsForScope();
    });
  };

  const handleSaveGroupInlinePolicy = async () => {
    await runAction(async () => {
      const groupId = Number.parseInt(selectedGroupId().trim(), 10);
      const statements = parseJSONArray<PolicyStatement>(groupInlinePolicyJson(), 'Group inline policy');
      const result = await putGroupInlinePolicy(
        groupId,
        groupInlinePolicyName().trim(),
        groupInlinePolicyDescription().trim(),
        statements
      );
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Saved group inline policy ${groupInlinePolicyName().trim()}.`);
      setGroupInlinePolicyName('');
      setGroupInlinePolicyDescription('');
      await loadSelectedGroup();
    });
  };

  const handleDeleteGroupInlinePolicy = async (name: string) => {
    await runAction(async () => {
      const groupId = Number.parseInt(selectedGroupId().trim(), 10);
      const result = await deleteGroupInlinePolicy(groupId, name);
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Deleted group inline policy ${name}.`);
      await loadSelectedGroup();
    });
  };

  const handleSaveTrustPolicy = async () => {
    await runAction(async () => {
      const statements = parseJSONArray<RoleTrustStatement>(trustJson(), 'Trust policy');
      const result = await putRoleTrustPolicy({
        roleName: trustRoleName().trim(),
        statements,
      });
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Saved trust policy for role ${trustRoleName().trim()}.`);
      await loadTrustPolicy();
    });
  };

  const handleSaveRoleBoundary = async () => {
    await runAction(async () => {
      const result = await putRolePermissionBoundary(
        trustRoleName().trim(),
        trustBoundaryPolicyName().trim()
      );
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Saved role boundary for ${trustRoleName().trim()}.`);
      await loadTrustPolicy();
    });
  };

  const handleDeleteRoleBoundary = async () => {
    await runAction(async () => {
      const result = await deleteRolePermissionBoundary(trustRoleName().trim());
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Deleted role boundary for ${trustRoleName().trim()}.`);
      setRoleBoundary(null);
      setTrustBoundaryPolicyName('');
    });
  };

  const handleSimulate = async () => {
    await runAction(async () => {
      const result = await simulateAccess({
        scope: simScope(),
        tenantId: simTenantId().trim() || undefined,
        userId: Number.parseInt(simTargetUserId().trim(), 10),
        action: simAction().trim(),
        resource: simResource().trim(),
        useAssumedRole: Boolean(buildAssumedRoleSession()),
        assumedRoleSession: buildAssumedRoleSession(),
        sessionPolicy: parseJSONArray<PolicyStatement>(simSessionPolicyJson(), 'Session policy'),
        attributes: parseJSONObject(simAttributesJson(), 'Simulation attributes'),
        servicePrincipal: simServicePrincipal().trim() || undefined,
        sessionTags: parseJSONObject(simSessionTagsJson(), 'Session tags'),
      });
      if (!result.success) throw new Error(result.message);
      setSimulation(result.data);
      setPageMessage(
        `Simulation completed: ${result.data.allowed ? 'allowed' : 'denied'} via ${result.data.decisionSource}.`
      );
    });
  };

  const handleAttachPrincipalManagedPolicy = async () => {
    await runAction(async () => {
      if (principalMode() === 'platform') {
        const targetUserId = Number.parseInt(principalPlatformUserId().trim(), 10);
        const result = await attachPlatformUserPolicy(targetUserId, principalManagedPolicyName().trim());
        if (!result.success) throw new Error(result.message);
      } else {
        const targetUserId = Number.parseInt(principalTenantUserId().trim(), 10);
        const result = await attachTenantUserPolicy(
          principalTenantId().trim(),
          targetUserId,
          principalManagedPolicyName().trim()
        );
        if (!result.success) throw new Error(result.message);
      }
      setPageMessage(`Attached managed policy ${principalManagedPolicyName().trim()}.`);
      setPrincipalManagedPolicyName('');
      await loadPrincipalControls();
      await loadSelectedPolicy();
    });
  };

  const handleDetachPrincipalManagedPolicy = async (policyName: string) => {
    await runAction(async () => {
      if (principalMode() === 'platform') {
        const targetUserId = Number.parseInt(principalPlatformUserId().trim(), 10);
        const result = await detachPlatformUserPolicy(targetUserId, policyName);
        if (!result.success) throw new Error(result.message);
      } else {
        const targetUserId = Number.parseInt(principalTenantUserId().trim(), 10);
        const result = await detachTenantUserPolicy(principalTenantId().trim(), targetUserId, policyName);
        if (!result.success) throw new Error(result.message);
      }
      setPageMessage(`Detached managed policy ${policyName}.`);
      await loadPrincipalControls();
      await loadSelectedPolicy();
    });
  };

  const handleSavePrincipalInlinePolicy = async () => {
    await runAction(async () => {
      const statements = parseJSONArray<PolicyStatement>(principalInlinePolicyJson(), 'Inline policy');
      if (principalMode() === 'platform') {
        const targetUserId = Number.parseInt(principalPlatformUserId().trim(), 10);
        const result = await putPlatformUserInlinePolicy(
          targetUserId,
          principalInlinePolicyName().trim(),
          principalInlinePolicyDescription().trim(),
          statements
        );
        if (!result.success) throw new Error(result.message);
      } else {
        const targetUserId = Number.parseInt(principalTenantUserId().trim(), 10);
        const result = await putTenantUserInlinePolicy(
          principalTenantId().trim(),
          targetUserId,
          principalInlinePolicyName().trim(),
          principalInlinePolicyDescription().trim(),
          statements
        );
        if (!result.success) throw new Error(result.message);
      }
      setPageMessage(`Saved inline policy ${principalInlinePolicyName().trim()}.`);
      setPrincipalInlinePolicyName('');
      setPrincipalInlinePolicyDescription('');
      await loadPrincipalControls();
    });
  };

  const handleDeletePrincipalInlinePolicy = async (name: string) => {
    await runAction(async () => {
      if (principalMode() === 'platform') {
        const targetUserId = Number.parseInt(principalPlatformUserId().trim(), 10);
        const result = await deletePlatformUserInlinePolicy(targetUserId, name);
        if (!result.success) throw new Error(result.message);
      } else {
        const targetUserId = Number.parseInt(principalTenantUserId().trim(), 10);
        const result = await deleteTenantUserInlinePolicy(principalTenantId().trim(), targetUserId, name);
        if (!result.success) throw new Error(result.message);
      }
      setPageMessage(`Deleted inline policy ${name}.`);
      await loadPrincipalControls();
    });
  };

  const handleSavePrincipalBoundary = async () => {
    await runAction(async () => {
      if (principalMode() === 'platform') {
        const targetUserId = Number.parseInt(principalPlatformUserId().trim(), 10);
        const result = await putPlatformUserPermissionBoundary(
          targetUserId,
          principalBoundaryPolicyName().trim()
        );
        if (!result.success) throw new Error(result.message);
      } else {
        const targetUserId = Number.parseInt(principalTenantUserId().trim(), 10);
        const result = await putTenantUserPermissionBoundary(
          principalTenantId().trim(),
          targetUserId,
          principalBoundaryPolicyName().trim()
        );
        if (!result.success) throw new Error(result.message);
      }
      setPageMessage(`Saved principal boundary ${principalBoundaryPolicyName().trim()}.`);
      await loadPrincipalControls();
      await loadSelectedPolicy();
    });
  };

  const handleDeletePrincipalBoundary = async () => {
    await runAction(async () => {
      if (principalMode() === 'platform') {
        const targetUserId = Number.parseInt(principalPlatformUserId().trim(), 10);
        const result = await deletePlatformUserPermissionBoundary(targetUserId);
        if (!result.success) throw new Error(result.message);
      } else {
        const targetUserId = Number.parseInt(principalTenantUserId().trim(), 10);
        const result = await deleteTenantUserPermissionBoundary(principalTenantId().trim(), targetUserId);
        if (!result.success) throw new Error(result.message);
      }
      setPageMessage('Deleted principal permission boundary.');
      setPrincipalBoundaryPolicyName('');
      await loadPrincipalControls();
      await loadSelectedPolicy();
    });
  };

  const handleAssignPlatformRole = async () => {
    await runAction(async () => {
      const targetUserId = Number.parseInt(shortcutPlatformUserId().trim(), 10);
      const result = await upsertPlatformRole({
        targetUserId,
        roleName: shortcutPlatformRoleName(),
      });
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Assigned platform role ${shortcutPlatformRoleName()} to user ${targetUserId}.`);
    });
  };

  const handleRemovePlatformRoleShortcut = async () => {
    await runAction(async () => {
      const targetUserId = Number.parseInt(shortcutPlatformUserId().trim(), 10);
      const result = await removePlatformRole(targetUserId, shortcutPlatformRoleName());
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Removed platform role ${shortcutPlatformRoleName()} from user ${targetUserId}.`);
    });
  };

  const handleAssignTenantRole = async () => {
    await runAction(async () => {
      const targetUserId = Number.parseInt(shortcutTenantUserId().trim(), 10);
      const result = await upsertTenantMember({
        tenantId: shortcutTenantId().trim(),
        userId: targetUserId,
        roleName: shortcutTenantRoleName(),
      });
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Assigned tenant role ${shortcutTenantRoleName()} to user ${targetUserId}.`);
    });
  };

  const handleRemoveTenantMembershipShortcut = async () => {
    await runAction(async () => {
      const targetUserId = Number.parseInt(shortcutTenantUserId().trim(), 10);
      const result = await removeTenantMember(shortcutTenantId().trim(), targetUserId);
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Removed tenant membership for user ${targetUserId}.`);
    });
  };

  const policyContextValue = {
    policyScopeOptions,
    policyScope,
    setPolicyScope,
    policyName,
    setPolicyName,
    policyDescription,
    setPolicyDescription,
    policyStatementsJson,
    setPolicyStatementsJson,
    policyVersionJson,
    setPolicyVersionJson,
    selectedPolicyName,
    setSelectedPolicyName,
    policyOptions,
    policyDetail,
    policyVersions,
    policyAttachments,
    attachmentColor,
    submitCreatePolicy,
    handleCreatePolicyVersion,
    handleDeletePolicy,
    handleSetDefaultVersion,
    handleDeleteVersion,
  };

  const groupContextValue = {
    groupScopeOptions,
    groupScope,
    setGroupScope,
    groupTenantId,
    setGroupTenantId,
    tenantOptions,
    groupName,
    setGroupName,
    groupDescription,
    setGroupDescription,
    submitCreateGroup,
    groupOptions,
    selectedGroupId,
    setSelectedGroupId,
    groupMemberUserId,
    setGroupMemberUserId,
    groupPolicyName,
    setGroupPolicyName,
    handleAddGroupMember,
    handleAttachGroupPolicy,
    handleDeleteGroup,
    groupMembers,
    handleRemoveGroupMember,
    groupPolicies,
    handleDetachGroupPolicy,
    groupInlinePolicyName,
    setGroupInlinePolicyName,
    groupInlinePolicyDescription,
    setGroupInlinePolicyDescription,
    groupInlinePolicyJson,
    setGroupInlinePolicyJson,
    handleSaveGroupInlinePolicy,
    groupInlinePolicies,
    handleDeleteGroupInlinePolicy,
  };

  const principalContextValue = {
    principalMode,
    setPrincipalMode,
    principalPlatformUserId,
    setPrincipalPlatformUserId,
    principalTenantId,
    setPrincipalTenantId,
    principalTenantUserId,
    setPrincipalTenantUserId,
    tenantOptions,
    principalManagedPolicyName,
    setPrincipalManagedPolicyName,
    handleAttachPrincipalManagedPolicy,
    currentManagedPolicies: () =>
      principalMode() === 'platform'
        ? platformUserPolicies()
        : tenantUserPolicies(),
    handleDetachPrincipalManagedPolicy,
    principalBoundaryPolicyName,
    setPrincipalBoundaryPolicyName,
    handleSavePrincipalBoundary,
    handleDeletePrincipalBoundary,
    currentBoundary: () =>
      principalMode() === 'platform'
        ? platformUserBoundary()
        : tenantUserBoundary(),
    principalInlinePolicyName,
    setPrincipalInlinePolicyName,
    principalInlinePolicyDescription,
    setPrincipalInlinePolicyDescription,
    principalInlinePolicyJson,
    setPrincipalInlinePolicyJson,
    handleSavePrincipalInlinePolicy,
    currentInlinePolicies: () =>
      principalMode() === 'platform'
        ? platformUserInlinePolicies()
        : tenantUserInlinePolicies(),
    handleDeletePrincipalInlinePolicy,
  };

  const trustSimContextValue = {
    trustRoleName,
    setTrustRoleName,
    loadTrustPolicy,
    handleSaveTrustPolicy,
    trustBoundaryPolicyName,
    setTrustBoundaryPolicyName,
    handleSaveRoleBoundary,
    handleDeleteRoleBoundary,
    roleBoundary,
    trustJson,
    setTrustJson,
    simScope,
    setSimScope,
    policyScopeOptions,
    simTargetUserId,
    setSimTargetUserId,
    tenantOptions,
    simTenantId,
    setSimTenantId,
    simAction,
    setSimAction,
    simResource,
    setSimResource,
    simServicePrincipal,
    setSimServicePrincipal,
    applyServiceAssumePreset,
    applyTenantAssumePreset,
    applyScopeDownDenyPreset,
    simAttributesJson,
    setSimAttributesJson,
    simSessionTagsJson,
    setSimSessionTagsJson,
    simSessionPolicyJson,
    setSimSessionPolicyJson,
    simAssumedRoleId,
    setSimAssumedRoleId,
    simAssumedRoleScope,
    setSimAssumedRoleScope,
    simAssumedRoleName,
    setSimAssumedRoleName,
    simAssumedRoleTenantId,
    setSimAssumedRoleTenantId,
    simAssumedRoleSessionName,
    setSimAssumedRoleSessionName,
    simAssumedRoleSourceIdentity,
    setSimAssumedRoleSourceIdentity,
    simAssumedRoleServicePrincipal,
    setSimAssumedRoleServicePrincipal,
    simAssumedRoleExpiresAt,
    setSimAssumedRoleExpiresAt,
    handleSimulate,
    simulation,
    simulationSourceColor,
    statementSourceLabel,
    simulationLayerTone,
  };

  return (
    <PageShell>
      <Card class="space-y-4">
        <SectionLead
          eyebrow="IAM Console"
          title="Operate AWS-style IAM controls for organizations, policies, groups, and session evaluation."
          copy="This console exposes the advanced surface behind the auth and IAM services: policy versioning, SCP governance, group bindings, trust policies, and access simulation."
        />
        <div class="flex flex-wrap gap-3">
          <Button href="/admin/settings" color="alternative" size="sm">
            Back to admin settings
          </Button>
          <Button href="/admin" color="light" size="sm">
            Back to admin home
          </Button>
        </div>
      </Card>

      <Show when={pageError()}>
        <ErrorAlert>{pageError()}</ErrorAlert>
      </Show>
      <Show when={pageMessage()}>
        <InfoAlert>{pageMessage()}</InfoAlert>
      </Show>
      <Show when={loading()}>
        <LoadingInline label="Loading IAM control plane..." />
      </Show>
      <Show when={!loading() && !allowed()}>
        <EmptyBlock
          title="IAM console unavailable"
          copy="This session does not have the platform permission required to manage advanced IAM controls."
        />
      </Show>

      <Show when={allowed()}>
        <div class="grid gap-6 xl:grid-cols-2">
          <Card class="space-y-4">
            <SectionTitle
              title="Organizations and SCP"
              subtitle="Create organizations, map workspaces, and attach service control policies."
            />
            <form class="grid gap-3 md:grid-cols-2" onSubmit={submitCreateOrganization}>
              <InputField label="Organization name" value={orgName()} onInput={(e) => setOrgName(e.currentTarget.value)} />
              <InputField label="Organization slug" value={orgSlug()} onInput={(e) => setOrgSlug(e.currentTarget.value)} />
              <div class="md:col-span-2">
                <Button type="submit" size="sm">Create organization</Button>
              </div>
            </form>

            <Show when={organizationOptions().length > 0}>
              <SelectField
                label="Selected organization"
                value={selectedOrgId()}
                options={organizationOptions()}
                onChange={(e) => setSelectedOrgId(e.currentTarget.value)}
              />
            </Show>

            <div class="grid gap-3 md:grid-cols-2">
              <SelectField
                label="Workspace to attach"
                value={orgTenantId()}
                options={tenantOptions()}
                onChange={(e) => setOrgTenantId(e.currentTarget.value)}
              />
              <InputField
                label="SCP policy name"
                value={orgPolicyName()}
                onInput={(e) => setOrgPolicyName(e.currentTarget.value)}
              />
            </div>

            <div class="flex flex-wrap gap-3">
              <Button size="sm" onClick={handleAttachTenantToOrg} disabled={!selectedOrgId() || !orgTenantId()}>
                Attach workspace
              </Button>
              <Button size="sm" color="dark" onClick={handleAttachScp} disabled={!selectedOrgId() || !orgPolicyName().trim()}>
                Attach SCP
              </Button>
            </div>

            <Show
              when={organizations().length > 0}
              fallback={<EmptyBlock title="No organizations" copy="Create the first organization to start applying SCP guardrails." />}
            >
              <div class="space-y-3">
                <For each={organizations()}>
                  {(org) => (
                    <div class="rounded-lg border border-gray-200 p-4">
                      <div class="flex flex-wrap items-center justify-between gap-3">
                        <div>
                          <p class="font-semibold text-gray-900">{org.name}</p>
                          <p class="text-sm text-gray-500">{org.slug}</p>
                        </div>
                        <Badge
                          content={org.id === selectedOrgId() ? 'selected' : 'organization'}
                          color={org.id === selectedOrgId() ? 'blue' : 'dark'}
                        />
                      </div>
                      <div class="mt-3 flex flex-wrap gap-2">
                        <Show when={org.id === selectedOrgId() && orgTenantId().trim()}>
                          <Button
                            size="xs"
                            color="light"
                            onClick={() => handleDetachTenantFromOrg(orgTenantId().trim())}
                          >
                            Detach selected workspace
                          </Button>
                        </Show>
                        <For each={org.id === selectedOrgId() ? orgPolicies() : []}>
                          {(policy) => (
                            <Button
                              size="xs"
                              color="alternative"
                              onClick={() => handleDetachScp(policy.name)}
                            >
                              Detach SCP {policy.name}
                            </Button>
                          )}
                        </For>
                      </div>
                    </div>
                  )}
                </For>
              </div>
            </Show>
          </Card>

          <Card class="space-y-4">
            <AdminIamPolicyProvider value={policyContextValue}>
              <PoliciesPanel />
            </AdminIamPolicyProvider>
          </Card>

          <Card class="space-y-4">
            <AdminIamGroupProvider value={groupContextValue}>
              <GroupsPanel />
            </AdminIamGroupProvider>
          </Card>

          <Card class="space-y-4">
            <SectionTitle
              title="Role assignment shortcuts"
              subtitle="Quickly grant or revoke platform roles and workspace memberships without leaving the IAM console."
            />
            <div class="grid gap-6 lg:grid-cols-2">
              <div class="space-y-3 rounded-lg border border-gray-200 p-4">
                <p class="text-sm font-semibold text-gray-900">Platform role shortcut</p>
                <InputField
                  label="Target user id"
                  value={shortcutPlatformUserId()}
                  onInput={(e) => setShortcutPlatformUserId(e.currentTarget.value)}
                />
                <SelectField
                  label="Platform role"
                  value={shortcutPlatformRoleName()}
                  options={platformRoleOptions}
                  onChange={(e) => setShortcutPlatformRoleName(e.currentTarget.value)}
                />
                <div class="flex flex-wrap gap-3">
                  <Button size="sm" onClick={handleAssignPlatformRole} disabled={!shortcutPlatformUserId().trim()}>
                    Assign platform role
                  </Button>
                  <Button size="sm" color="red" onClick={handleRemovePlatformRoleShortcut} disabled={!shortcutPlatformUserId().trim()}>
                    Remove platform role
                  </Button>
                </div>
              </div>

              <div class="space-y-3 rounded-lg border border-gray-200 p-4">
                <p class="text-sm font-semibold text-gray-900">Workspace membership shortcut</p>
                <SelectField
                  label="Workspace"
                  value={shortcutTenantId()}
                  options={tenantOptions()}
                  onChange={(e) => setShortcutTenantId(e.currentTarget.value)}
                />
                <InputField
                  label="Target user id"
                  value={shortcutTenantUserId()}
                  onInput={(e) => setShortcutTenantUserId(e.currentTarget.value)}
                />
                <SelectField
                  label="Workspace role"
                  value={shortcutTenantRoleName()}
                  options={tenantRoleOptions}
                  onChange={(e) => setShortcutTenantRoleName(e.currentTarget.value)}
                />
                <div class="flex flex-wrap gap-3">
                  <Button size="sm" onClick={handleAssignTenantRole} disabled={!shortcutTenantId().trim() || !shortcutTenantUserId().trim()}>
                    Assign workspace role
                  </Button>
                  <Button size="sm" color="red" onClick={handleRemoveTenantMembershipShortcut} disabled={!shortcutTenantId().trim() || !shortcutTenantUserId().trim()}>
                    Remove membership
                  </Button>
                </div>
              </div>
            </div>
          </Card>

          <Card class="space-y-4">
            <AdminIamPrincipalProvider value={principalContextValue}>
              <PrincipalPoliciesPanel />
            </AdminIamPrincipalProvider>
          </Card>

          <Card class="space-y-4">
            <AdminIamTrustSimProvider value={trustSimContextValue}>
              <TrustSimulationPanel />
            </AdminIamTrustSimProvider>
          </Card>
        </div>
      </Show>
    </PageShell>
  );
}
