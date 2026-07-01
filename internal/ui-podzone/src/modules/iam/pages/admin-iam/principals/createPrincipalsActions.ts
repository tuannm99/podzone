import {
  attachPlatformUserPolicy,
  attachTenantUserPolicy,
  deletePlatformUserInlinePolicy,
  deletePlatformUserPermissionBoundary,
  deleteTenantUserInlinePolicy,
  deleteTenantUserPermissionBoundary,
  detachPlatformUserPolicy,
  detachTenantUserPolicy,
  putPlatformUserInlinePolicy,
  putPlatformUserPermissionBoundary,
  putTenantUserInlinePolicy,
  putTenantUserPermissionBoundary,
  type PolicyStatement,
} from '@/services/iam'
import type { AdminIamLoaders } from '../createAdminIamLoaders'
import type { AdminIamState } from '../createAdminIamState'
import { parseJSONArray } from '../presentation'
import type { RunAction } from '../shared/actions'
import type {
  PrincipalBoundaryFormValues,
  PrincipalInlinePolicyFormValues,
  PrincipalManagedPolicyFormValues,
} from './forms'

export function createPrincipalsActions(
  state: AdminIamState,
  loaders: AdminIamLoaders,
  runAction: RunAction
) {
  const userID = () =>
    Number.parseInt(
      state.principalMode() === 'platform'
        ? state.principalPlatformUserId().trim()
        : state.principalTenantUserId().trim(),
      10
    )

  const attachPrincipalManagedPolicyFromForm = (
    values: PrincipalManagedPolicyFormValues
  ) =>
    runAction(async () => {
      const policyName = values.policyName.trim()
      const result =
        state.principalMode() === 'platform'
          ? await attachPlatformUserPolicy(userID(), policyName)
          : await attachTenantUserPolicy(
              state.principalTenantId().trim(),
              userID(),
              policyName
            )
      if (!result.success) throw new Error(result.message)
      state.setPageMessage(`Attached managed policy ${policyName}.`)
      state.setPrincipalManagedPolicyName('')
      await loaders.loadPrincipalControls()
      await loaders.loadSelectedPolicy()
    })

  const handleAttachPrincipalManagedPolicy = () =>
    attachPrincipalManagedPolicyFromForm({
      policyName: state.principalManagedPolicyName(),
    })

  const handleDetachPrincipalManagedPolicy = (policyName: string) =>
    runAction(async () => {
      const result =
        state.principalMode() === 'platform'
          ? await detachPlatformUserPolicy(userID(), policyName)
          : await detachTenantUserPolicy(
              state.principalTenantId().trim(),
              userID(),
              policyName
            )
      if (!result.success) throw new Error(result.message)
      state.setPageMessage(`Detached managed policy ${policyName}.`)
      await loaders.loadPrincipalControls()
      await loaders.loadSelectedPolicy()
    })

  const savePrincipalInlinePolicyFromForm = (
    values: PrincipalInlinePolicyFormValues
  ) =>
    runAction(async () => {
      const statements = parseJSONArray<PolicyStatement>(
        values.statementsJson,
        'Inline policy'
      )
      const result =
        state.principalMode() === 'platform'
          ? await putPlatformUserInlinePolicy(
              userID(),
              values.name.trim(),
              values.description.trim(),
              statements
            )
          : await putTenantUserInlinePolicy(
              state.principalTenantId().trim(),
              userID(),
              values.name.trim(),
              values.description.trim(),
              statements
            )
      if (!result.success) throw new Error(result.message)
      state.setPageMessage(`Saved inline policy ${values.name.trim()}.`)
      state.setPrincipalInlinePolicyName('')
      state.setPrincipalInlinePolicyDescription('')
      state.setPrincipalInlinePolicyJson(values.statementsJson)
      await loaders.loadPrincipalControls()
    })

  const handleSavePrincipalInlinePolicy = () =>
    savePrincipalInlinePolicyFromForm({
      name: state.principalInlinePolicyName(),
      description: state.principalInlinePolicyDescription(),
      statementsJson: state.principalInlinePolicyJson(),
    })

  const handleDeletePrincipalInlinePolicy = (name: string) =>
    runAction(async () => {
      const result =
        state.principalMode() === 'platform'
          ? await deletePlatformUserInlinePolicy(userID(), name)
          : await deleteTenantUserInlinePolicy(
              state.principalTenantId().trim(),
              userID(),
              name
            )
      if (!result.success) throw new Error(result.message)
      state.setPageMessage(`Deleted inline policy ${name}.`)
      await loaders.loadPrincipalControls()
    })

  const savePrincipalBoundaryFromForm = (values: PrincipalBoundaryFormValues) =>
    runAction(async () => {
      const policyName = values.policyName.trim()
      const result =
        state.principalMode() === 'platform'
          ? await putPlatformUserPermissionBoundary(userID(), policyName)
          : await putTenantUserPermissionBoundary(
              state.principalTenantId().trim(),
              userID(),
              policyName
            )
      if (!result.success) throw new Error(result.message)
      state.setPageMessage(`Saved principal boundary ${policyName}.`)
      state.setPrincipalBoundaryPolicyName(policyName)
      await loaders.loadPrincipalControls()
      await loaders.loadSelectedPolicy()
    })

  const handleSavePrincipalBoundary = () =>
    savePrincipalBoundaryFromForm({
      policyName: state.principalBoundaryPolicyName(),
    })

  const handleDeletePrincipalBoundary = () =>
    runAction(async () => {
      const result =
        state.principalMode() === 'platform'
          ? await deletePlatformUserPermissionBoundary(userID())
          : await deleteTenantUserPermissionBoundary(
              state.principalTenantId().trim(),
              userID()
            )
      if (!result.success) throw new Error(result.message)
      state.setPageMessage('Deleted principal permission boundary.')
      state.setPrincipalBoundaryPolicyName('')
      await loaders.loadPrincipalControls()
      await loaders.loadSelectedPolicy()
    })

  return {
    attachPrincipalManagedPolicyFromForm,
    handleAttachPrincipalManagedPolicy,
    handleDetachPrincipalManagedPolicy,
    savePrincipalInlinePolicyFromForm,
    handleSavePrincipalInlinePolicy,
    handleDeletePrincipalInlinePolicy,
    savePrincipalBoundaryFromForm,
    handleSavePrincipalBoundary,
    handleDeletePrincipalBoundary,
  }
}
