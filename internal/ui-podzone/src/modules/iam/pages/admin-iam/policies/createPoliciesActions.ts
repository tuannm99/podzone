import {
  createPolicy,
  createPolicyVersion,
  deletePolicy,
  deletePolicyVersion,
  setDefaultPolicyVersion,
  type PolicyStatement,
} from '@/services/iam'
import type { AdminIamLoaders } from '../createAdminIamLoaders'
import type { AdminIamState } from '../createAdminIamState'
import { parseJSONArray } from '../presentation'
import type { RunAction } from '../shared/actions'
import type {
  CreatePolicyFormValues,
  CreatePolicyVersionFormValues,
} from './forms'

export function createPoliciesActions(
  state: AdminIamState,
  loaders: AdminIamLoaders,
  runAction: RunAction
) {
  const createPolicyFromForm = (values: CreatePolicyFormValues) =>
    runAction(async () => {
      const statements = parseJSONArray<PolicyStatement>(
        values.statementsJson,
        'Policy statements'
      )
      const result = await createPolicy({
        scope: values.scope,
        name: values.name.trim(),
        description: values.description.trim(),
        statements,
      })
      if (!result.success) throw new Error(result.message)
      state.setPageMessage(`Created policy ${values.name.trim()}.`)
      state.setPolicyName('')
      state.setPolicyDescription('')
      state.setPolicyScope(values.scope)
      state.setPolicyStatementsJson(values.statementsJson)
      await state.reloadPolicies()
    })

  const submitCreatePolicy = async (event: SubmitEvent) => {
    event.preventDefault()
    await createPolicyFromForm({
      scope: state.policyScope(),
      name: state.policyName(),
      description: state.policyDescription(),
      statementsJson: state.policyStatementsJson(),
    })
  }

  const createPolicyVersionFromForm = (values: CreatePolicyVersionFormValues) =>
    runAction(async () => {
      const statements = parseJSONArray<PolicyStatement>(
        values.statementsJson,
        'Policy version statements'
      )
      const result = await createPolicyVersion({
        name: state.selectedPolicyName().trim(),
        statements,
        setAsDefault: false,
      })
      if (!result.success) throw new Error(result.message)
      state.setPolicyVersionJson(values.statementsJson)
      state.setPageMessage(
        `Created a new version for ${state.selectedPolicyName().trim()}.`
      )
      await loaders.loadSelectedPolicy()
    })

  const handleCreatePolicyVersion = () =>
    createPolicyVersionFromForm({
      statementsJson: state.policyVersionJson(),
    })

  const handleDeletePolicy = () =>
    runAction(async () => {
      const name = state.selectedPolicyName().trim()
      const result = await deletePolicy(name)
      if (!result.success) throw new Error(result.message)
      state.setPageMessage(`Deleted policy ${name}.`)
      state.setPolicyDetail(undefined)
      state.setPolicyVersions([])
      state.setPolicyAttachments([])
      await state.reloadPolicies()
      state.setSelectedPolicyName(state.policies()[0]?.name || '')
    })

  const handleSetDefaultVersion = (version: string) =>
    runAction(async () => {
      const policyName = state.selectedPolicyName().trim()
      const result = await setDefaultPolicyVersion(policyName, version)
      if (!result.success) throw new Error(result.message)
      state.setPageMessage(`Set ${version} as default for ${policyName}.`)
      await loaders.loadSelectedPolicy()
    })

  const handleDeleteVersion = (version: string) =>
    runAction(async () => {
      const result = await deletePolicyVersion(
        state.selectedPolicyName().trim(),
        version
      )
      if (!result.success) throw new Error(result.message)
      state.setPageMessage(`Deleted policy version ${version}.`)
      await loaders.loadSelectedPolicy()
    })

  return {
    submitCreatePolicy,
    createPolicyFromForm,
    createPolicyVersionFromForm,
    handleCreatePolicyVersion,
    handleDeletePolicy,
    handleSetDefaultVersion,
    handleDeleteVersion,
  }
}
