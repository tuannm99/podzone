import { createEffect } from 'solid-js';
import {
  attachPlatformUserPolicy,
  attachTenantUserPolicy,
  deletePlatformUserInlinePolicy,
  deletePlatformUserPermissionBoundary,
  deleteRolePermissionBoundary,
  deleteTenantUserInlinePolicy,
  deleteTenantUserPermissionBoundary,
  detachPlatformUserPolicy,
  detachTenantUserPolicy,
  putPlatformUserInlinePolicy,
  putPlatformUserPermissionBoundary,
  putRolePermissionBoundary,
  putRoleTrustPolicy,
  putTenantUserInlinePolicy,
  putTenantUserPermissionBoundary,
  removePlatformRole,
  removeTenantMember,
  simulateAccess,
  type PolicyStatement,
  type RoleTrustStatement,
  upsertPlatformRole,
  upsertTenantMember,
} from '@/services/iam';
import { parseJSONArray, parseJSONObject, prettyJSON } from './presentation';
import type {
  PrincipalBoundaryFormValues,
  PrincipalInlinePolicyFormValues,
  PrincipalManagedPolicyFormValues,
} from './principal-forms';
import type { AdminIamLoaders } from './useAdminIamLoaders';
import type { AdminIamState } from './useAdminIamState';

type RunAction = (work: () => Promise<void>) => Promise<void>;

export function useAdminIamPrincipalTrustActions(state: AdminIamState, loaders: AdminIamLoaders, runAction: RunAction) {
  const buildAssumedRoleSession = () => {
    const roleId = Number.parseInt(state.simAssumedRoleId().trim(), 10);
    if (!Number.isFinite(roleId) || roleId <= 0) return undefined;
    return {
      assumedRoleId: roleId,
      assumedRoleScope: state.simAssumedRoleScope().trim(),
      assumedRoleName: state.simAssumedRoleName().trim(),
      assumedRoleTenantId: state.simAssumedRoleTenantId().trim() || undefined,
      assumedRoleSessionName: state.simAssumedRoleSessionName().trim() || undefined,
      assumedRoleSourceIdentity: state.simAssumedRoleSourceIdentity().trim() || undefined,
      assumedRoleServicePrincipal: state.simAssumedRoleServicePrincipal().trim() || undefined,
      assumedRoleExpiresAt: state.simAssumedRoleExpiresAt().trim() || undefined,
      sessionTags: parseJSONObject(state.simSessionTagsJson(), 'Session tags'),
    };
  };

  createEffect(() => {
    if (
      state.simAssumedRoleScope() === 'tenant' &&
      state.simTenantId().trim() &&
      !state.simAssumedRoleTenantId().trim()
    ) {
      state.setSimAssumedRoleTenantId(state.simTenantId().trim());
    }
  });

  const applyServiceAssumePreset = () => {
    state.setSimScope('platform');
    state.setSimAction('order:update');
    state.setSimResource('*');
    state.setSimServicePrincipal('backoffice.podzone.internal');
    state.setSimAssumedRoleScope('platform');
    state.setSimAssumedRoleName('platform_admin');
    state.setSimAssumedRoleSourceIdentity('backoffice-admin');
    state.setSimAssumedRoleSessionName('service-assume');
    state.setSimAssumedRoleServicePrincipal('backoffice.podzone.internal');
    state.setSimAttributesJson(prettyJSON({ lane: 'priority' }));
    state.setSimSessionTagsJson(
      prettyJSON({ team: 'ops', path: 'service-assume' })
    );
  };

  const applyTenantAssumePreset = () => {
    state.setSimScope('tenant');
    state.setSimAction('order:update');
    state.setSimResource('*');
    state.setSimAssumedRoleScope('tenant');
    state.setSimAssumedRoleName('tenant_admin');
    state.setSimAssumedRoleTenantId(state.simTenantId().trim());
    state.setSimAssumedRoleSessionName('tenant-admin-review');
    state.setSimAssumedRoleSourceIdentity('store-ops');
    state.setSimAssumedRoleServicePrincipal('');
    state.setSimAttributesJson(prettyJSON({ lane: 'priority', region: 'us' }));
    state.setSimSessionTagsJson(
      prettyJSON({
        team: 'ops',
        store: state.simTenantId().trim() || 'tenant',
      })
    );
  };

  const applyScopeDownDenyPreset = () => {
    state.setSimAction('order:update');
    state.setSimResource('*');
    state.setSimSessionPolicyJson(
      prettyJSON([
        {
          effect: 'deny',
          actionPattern: 'order:update',
          resourcePattern: '*',
          conditions: [],
        },
      ])
    );
    state.setSimAttributesJson(prettyJSON({ lane: 'restricted' }));
    state.setSimSessionTagsJson(prettyJSON({ team: 'ops', mode: 'scope-down' }));
  };

  const handleSaveTrustPolicy = async () => {
    await runAction(async () => {
      const statements = parseJSONArray<RoleTrustStatement>(
        state.trustJson(),
        'Trust policy'
      );
      const result = await putRoleTrustPolicy({
        roleName: state.trustRoleName().trim(),
        statements,
      });
      if (!result.success) throw new Error(result.message);
      state.setPageMessage(
        `Saved trust policy for role ${state.trustRoleName().trim()}.`
      );
      await loaders.loadTrustPolicy();
    });
  };

  const handleSaveRoleBoundary = async () => {
    await runAction(async () => {
      const result = await putRolePermissionBoundary(
        state.trustRoleName().trim(),
        state.trustBoundaryPolicyName().trim()
      );
      if (!result.success) throw new Error(result.message);
      state.setPageMessage(
        `Saved role boundary for ${state.trustRoleName().trim()}.`
      );
      await loaders.loadTrustPolicy();
    });
  };

  const handleDeleteRoleBoundary = async () => {
    await runAction(async () => {
      const result = await deleteRolePermissionBoundary(
        state.trustRoleName().trim()
      );
      if (!result.success) throw new Error(result.message);
      state.setPageMessage(
        `Deleted role boundary for ${state.trustRoleName().trim()}.`
      );
      state.setRoleBoundary(null);
      state.setTrustBoundaryPolicyName('');
    });
  };

  const handleSimulate = async () => {
    await runAction(async () => {
      const assumedRoleSession = buildAssumedRoleSession();
      const result = await simulateAccess({
        scope: state.simScope(),
        tenantId: state.simTenantId().trim() || undefined,
        userId: Number.parseInt(state.simTargetUserId().trim(), 10),
        action: state.simAction().trim(),
        resource: state.simResource().trim(),
        useAssumedRole: Boolean(assumedRoleSession),
        assumedRoleSession,
        sessionPolicy: parseJSONArray<PolicyStatement>(
          state.simSessionPolicyJson(),
          'Session policy'
        ),
        attributes: parseJSONObject(
          state.simAttributesJson(),
          'Simulation attributes'
        ),
        servicePrincipal: state.simServicePrincipal().trim() || undefined,
        sessionTags: parseJSONObject(state.simSessionTagsJson(), 'Session tags'),
      });
      if (!result.success) throw new Error(result.message);
      state.setSimulation(result.data);
      state.setPageMessage(
        `Simulation completed: ${result.data.allowed ? 'allowed' : 'denied'} via ${result.data.decisionSource}.`
      );
    });
  };

  const attachPrincipalManagedPolicyFromForm = async (
    values: PrincipalManagedPolicyFormValues
  ) => {
    await runAction(async () => {
      if (state.principalMode() === 'platform') {
        const targetUserId = Number.parseInt(
          state.principalPlatformUserId().trim(),
          10
        );
        const result = await attachPlatformUserPolicy(
          targetUserId,
          values.policyName.trim()
        );
        if (!result.success) throw new Error(result.message);
      } else {
        const targetUserId = Number.parseInt(
          state.principalTenantUserId().trim(),
          10
        );
        const result = await attachTenantUserPolicy(
          state.principalTenantId().trim(),
          targetUserId,
          values.policyName.trim()
        );
        if (!result.success) throw new Error(result.message);
      }
      state.setPageMessage(
        `Attached managed policy ${values.policyName.trim()}.`
      );
      state.setPrincipalManagedPolicyName('');
      await loaders.loadPrincipalControls();
      await loaders.loadSelectedPolicy();
    });
  };

  const handleAttachPrincipalManagedPolicy = async () => {
    await attachPrincipalManagedPolicyFromForm({
      policyName: state.principalManagedPolicyName(),
    });
  };

  const handleDetachPrincipalManagedPolicy = async (policyName: string) => {
    await runAction(async () => {
      if (state.principalMode() === 'platform') {
        const targetUserId = Number.parseInt(
          state.principalPlatformUserId().trim(),
          10
        );
        const result = await detachPlatformUserPolicy(targetUserId, policyName);
        if (!result.success) throw new Error(result.message);
      } else {
        const targetUserId = Number.parseInt(
          state.principalTenantUserId().trim(),
          10
        );
        const result = await detachTenantUserPolicy(
          state.principalTenantId().trim(),
          targetUserId,
          policyName
        );
        if (!result.success) throw new Error(result.message);
      }
      state.setPageMessage(`Detached managed policy ${policyName}.`);
      await loaders.loadPrincipalControls();
      await loaders.loadSelectedPolicy();
    });
  };

  const savePrincipalInlinePolicyFromForm = async (
    values: PrincipalInlinePolicyFormValues
  ) => {
    await runAction(async () => {
      const statements = parseJSONArray<PolicyStatement>(
        values.statementsJson,
        'Inline policy'
      );
      if (state.principalMode() === 'platform') {
        const targetUserId = Number.parseInt(
          state.principalPlatformUserId().trim(),
          10
        );
        const result = await putPlatformUserInlinePolicy(
          targetUserId,
          values.name.trim(),
          values.description.trim(),
          statements
        );
        if (!result.success) throw new Error(result.message);
      } else {
        const targetUserId = Number.parseInt(
          state.principalTenantUserId().trim(),
          10
        );
        const result = await putTenantUserInlinePolicy(
          state.principalTenantId().trim(),
          targetUserId,
          values.name.trim(),
          values.description.trim(),
          statements
        );
        if (!result.success) throw new Error(result.message);
      }
      state.setPageMessage(
        `Saved inline policy ${values.name.trim()}.`
      );
      state.setPrincipalInlinePolicyName('');
      state.setPrincipalInlinePolicyDescription('');
      state.setPrincipalInlinePolicyJson(values.statementsJson);
      await loaders.loadPrincipalControls();
    });
  };

  const handleSavePrincipalInlinePolicy = async () => {
    await savePrincipalInlinePolicyFromForm({
      name: state.principalInlinePolicyName(),
      description: state.principalInlinePolicyDescription(),
      statementsJson: state.principalInlinePolicyJson(),
    });
  };

  const handleDeletePrincipalInlinePolicy = async (name: string) => {
    await runAction(async () => {
      if (state.principalMode() === 'platform') {
        const targetUserId = Number.parseInt(
          state.principalPlatformUserId().trim(),
          10
        );
        const result = await deletePlatformUserInlinePolicy(targetUserId, name);
        if (!result.success) throw new Error(result.message);
      } else {
        const targetUserId = Number.parseInt(
          state.principalTenantUserId().trim(),
          10
        );
        const result = await deleteTenantUserInlinePolicy(
          state.principalTenantId().trim(),
          targetUserId,
          name
        );
        if (!result.success) throw new Error(result.message);
      }
      state.setPageMessage(`Deleted inline policy ${name}.`);
      await loaders.loadPrincipalControls();
    });
  };

  const savePrincipalBoundaryFromForm = async (
    values: PrincipalBoundaryFormValues
  ) => {
    await runAction(async () => {
      if (state.principalMode() === 'platform') {
        const targetUserId = Number.parseInt(
          state.principalPlatformUserId().trim(),
          10
        );
        const result = await putPlatformUserPermissionBoundary(
          targetUserId,
          values.policyName.trim()
        );
        if (!result.success) throw new Error(result.message);
      } else {
        const targetUserId = Number.parseInt(
          state.principalTenantUserId().trim(),
          10
        );
        const result = await putTenantUserPermissionBoundary(
          state.principalTenantId().trim(),
          targetUserId,
          values.policyName.trim()
        );
        if (!result.success) throw new Error(result.message);
      }
      state.setPageMessage(
        `Saved principal boundary ${values.policyName.trim()}.`
      );
      state.setPrincipalBoundaryPolicyName(values.policyName);
      await loaders.loadPrincipalControls();
      await loaders.loadSelectedPolicy();
    });
  };

  const handleSavePrincipalBoundary = async () => {
    await savePrincipalBoundaryFromForm({
      policyName: state.principalBoundaryPolicyName(),
    });
  };

  const handleDeletePrincipalBoundary = async () => {
    await runAction(async () => {
      if (state.principalMode() === 'platform') {
        const targetUserId = Number.parseInt(
          state.principalPlatformUserId().trim(),
          10
        );
        const result = await deletePlatformUserPermissionBoundary(targetUserId);
        if (!result.success) throw new Error(result.message);
      } else {
        const targetUserId = Number.parseInt(
          state.principalTenantUserId().trim(),
          10
        );
        const result = await deleteTenantUserPermissionBoundary(
          state.principalTenantId().trim(),
          targetUserId
        );
        if (!result.success) throw new Error(result.message);
      }
      state.setPageMessage('Deleted principal permission boundary.');
      state.setPrincipalBoundaryPolicyName('');
      await loaders.loadPrincipalControls();
      await loaders.loadSelectedPolicy();
    });
  };

  const handleAssignPlatformRole = async () => {
    await runAction(async () => {
      const targetUserId = Number.parseInt(
        state.shortcutPlatformUserId().trim(),
        10
      );
      const result = await upsertPlatformRole({
        targetUserId,
        roleName: state.shortcutPlatformRoleName(),
      });
      if (!result.success) throw new Error(result.message);
      state.setPageMessage(
        `Assigned platform role ${state.shortcutPlatformRoleName()} to user ${targetUserId}.`
      );
    });
  };

  const handleRemovePlatformRoleShortcut = async () => {
    await runAction(async () => {
      const targetUserId = Number.parseInt(
        state.shortcutPlatformUserId().trim(),
        10
      );
      const result = await removePlatformRole(
        targetUserId,
        state.shortcutPlatformRoleName()
      );
      if (!result.success) throw new Error(result.message);
      state.setPageMessage(
        `Removed platform role ${state.shortcutPlatformRoleName()} from user ${targetUserId}.`
      );
    });
  };

  const handleAssignTenantRole = async () => {
    await runAction(async () => {
      const targetUserId = Number.parseInt(
        state.shortcutTenantUserId().trim(),
        10
      );
      const result = await upsertTenantMember({
        tenantId: state.shortcutTenantId().trim(),
        userId: targetUserId,
        roleName: state.shortcutTenantRoleName(),
      });
      if (!result.success) throw new Error(result.message);
      state.setPageMessage(
        `Assigned tenant role ${state.shortcutTenantRoleName()} to user ${targetUserId}.`
      );
    });
  };

  const handleRemoveTenantMembershipShortcut = async () => {
    await runAction(async () => {
      const targetUserId = Number.parseInt(
        state.shortcutTenantUserId().trim(),
        10
      );
      const result = await removeTenantMember(
        state.shortcutTenantId().trim(),
        targetUserId
      );
      if (!result.success) throw new Error(result.message);
      state.setPageMessage(`Removed tenant membership for user ${targetUserId}.`);
    });
  };

  return {
    applyServiceAssumePreset,
    applyTenantAssumePreset,
    applyScopeDownDenyPreset,
    handleSaveTrustPolicy,
    handleSaveRoleBoundary,
    handleDeleteRoleBoundary,
    handleSimulate,
    attachPrincipalManagedPolicyFromForm,
    handleAttachPrincipalManagedPolicy,
    handleDetachPrincipalManagedPolicy,
    savePrincipalInlinePolicyFromForm,
    handleSavePrincipalInlinePolicy,
    handleDeletePrincipalInlinePolicy,
    savePrincipalBoundaryFromForm,
    handleSavePrincipalBoundary,
    handleDeletePrincipalBoundary,
    handleAssignPlatformRole,
    handleRemovePlatformRoleShortcut,
    handleAssignTenantRole,
    handleRemoveTenantMembershipShortcut,
  };
}
