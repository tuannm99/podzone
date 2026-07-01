import { createSignal } from 'solid-js'
import type { OrganizationInfo, PolicyInfo } from '@/services/iam'

export function createOrganizationsState() {
  const [organizations, setOrganizations] = createSignal<OrganizationInfo[]>([])
  const [selectedOrgId, setSelectedOrgId] = createSignal('')
  const [orgPolicies, setOrgPolicies] = createSignal<PolicyInfo[]>([])
  const [orgName, setOrgName] = createSignal('')
  const [orgSlug, setOrgSlug] = createSignal('')
  const [orgTenantId, setOrgTenantId] = createSignal('')
  const [orgPolicyName, setOrgPolicyName] = createSignal('')
  const organizationOptions = () =>
    organizations().map((item) => ({
      name: `${item.slug} · ${item.name}`,
      value: item.id,
    }))

  return {
    organizations,
    setOrganizations,
    selectedOrgId,
    setSelectedOrgId,
    orgPolicies,
    setOrgPolicies,
    orgName,
    setOrgName,
    orgSlug,
    setOrgSlug,
    orgTenantId,
    setOrgTenantId,
    orgPolicyName,
    setOrgPolicyName,
    organizationOptions,
  }
}
