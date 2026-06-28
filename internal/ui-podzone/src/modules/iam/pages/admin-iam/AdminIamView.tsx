import { Show, createSignal, onMount } from 'solid-js'
import type { AdminIamViewModel } from '../AdminIamPage'
import {
  EmptyBlock,
  ErrorAlert,
  InfoAlert,
  LoadingInline,
} from '@/solid/components/common/Feedback'
import { PageShell } from '@/solid/components/common/PageShell'
import { Button, Card } from '@/solid/components/common/Primitives'
import { SectionLead } from '@/solid/components/common/SectionLead'
import { AdminIamGroupProvider } from './group-context'
import { GroupsPanel } from './GroupsPanel'
import { IamWorkspaceNav } from './IamWorkspaceNav'
import { OrganizationsPanel } from './OrganizationsPanel'
import { AdminIamPolicyProvider } from './policy-context'
import { PoliciesPanel } from './PoliciesPanel'
import { type IamSection, type IamSectionID } from './presentation'
import { AdminIamPrincipalProvider } from './principal-context'
import { PrincipalPoliciesPanel } from './PrincipalPoliciesPanel'
import { RoleAssignmentsPanel } from './RoleAssignmentsPanel'
import { AdminIamTrustSimProvider } from './trust-sim-context'
import { TrustSimulationPanel } from './TrustSimulationPanel'

// The controller page assembles the model; this view only owns composition.
export function AdminIamView(props: { model: AdminIamViewModel }) {
  const {
    pageError,
    pageMessage,
    loading,
    allowed,
    sectionLinks,
    policyContextValue,
    groupContextValue,
    principalContextValue,
    trustSimContextValue,
    organizationOptions,
    selectedOrgId,
    setSelectedOrgId,
    submitCreateOrganization,
    orgName,
    setOrgName,
    orgSlug,
    setOrgSlug,
    orgTenantId,
    setOrgTenantId,
    orgPolicyName,
    setOrgPolicyName,
    tenantOptions,
    handleAttachTenantToOrg,
    handleAttachScp,
    organizations,
    orgPolicies,
    handleDetachTenantFromOrg,
    handleDetachScp,
    shortcutPlatformUserId,
    setShortcutPlatformUserId,
    shortcutPlatformRoleName,
    setShortcutPlatformRoleName,
    platformRoleOptions,
    handleAssignPlatformRole,
    handleRemovePlatformRoleShortcut,
    shortcutTenantId,
    setShortcutTenantId,
    shortcutTenantUserId,
    setShortcutTenantUserId,
    shortcutTenantRoleName,
    setShortcutTenantRoleName,
    tenantRoleOptions,
    handleAssignTenantRole,
    handleRemoveTenantMembershipShortcut,
  } = props.model

  const sections = sectionLinks as IamSection[]
  const [activeSection, setActiveSection] =
    createSignal<IamSectionID>('iam-policies')

  const selectSection = (section: IamSectionID) => {
    setActiveSection(section)
    window.history.replaceState(null, '', `#${section}`)
  }

  onMount(() => {
    const hash = window.location.hash.slice(1)
    const selected = sections.find((section) => section.id === hash)
    if (selected) setActiveSection(selected.id)
  })

  return (
    <PageShell>
      <header class="space-y-4 border-b border-gray-200 pb-5">
        <SectionLead
          eyebrow="IAM Console"
          title="IAM control plane"
          copy="Centralized policy, principal, organization, and access evaluation controls."
        />
        <div class="flex flex-wrap gap-3">
          <Button href="/admin/settings" color="alternative" size="sm">
            Back to admin settings
          </Button>
          <Button href="/admin" color="light" size="sm">
            Back to admin home
          </Button>
        </div>
      </header>

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
          copy="This session does not have the required platform permission."
        />
      </Show>

      <Show when={allowed()}>
        <IamWorkspaceNav
          sections={sections}
          activeSection={activeSection()}
          onSelect={selectSection}
        />

        <Card class="space-y-4">
          <Show when={activeSection() === 'iam-orgs'}>
            <OrganizationsPanel
              organizationOptions={organizationOptions}
              selectedOrgId={selectedOrgId}
              setSelectedOrgId={setSelectedOrgId}
              submitCreateOrganization={submitCreateOrganization}
              orgName={orgName}
              setOrgName={setOrgName}
              orgSlug={orgSlug}
              setOrgSlug={setOrgSlug}
              orgTenantId={orgTenantId}
              setOrgTenantId={setOrgTenantId}
              orgPolicyName={orgPolicyName}
              setOrgPolicyName={setOrgPolicyName}
              tenantOptions={tenantOptions}
              handleAttachTenantToOrg={handleAttachTenantToOrg}
              handleAttachScp={handleAttachScp}
              organizations={organizations}
              orgPolicies={orgPolicies}
              handleDetachTenantFromOrg={handleDetachTenantFromOrg}
              handleDetachScp={handleDetachScp}
            />
          </Show>

          <Show when={activeSection() === 'iam-policies'}>
            <AdminIamPolicyProvider value={policyContextValue}>
              <PoliciesPanel />
            </AdminIamPolicyProvider>
          </Show>

          <Show when={activeSection() === 'iam-groups'}>
            <AdminIamGroupProvider value={groupContextValue}>
              <GroupsPanel />
            </AdminIamGroupProvider>
          </Show>

          <Show when={activeSection() === 'iam-assignments'}>
            <RoleAssignmentsPanel
              shortcutPlatformUserId={shortcutPlatformUserId}
              setShortcutPlatformUserId={setShortcutPlatformUserId}
              shortcutPlatformRoleName={shortcutPlatformRoleName}
              setShortcutPlatformRoleName={setShortcutPlatformRoleName}
              platformRoleOptions={platformRoleOptions}
              handleAssignPlatformRole={handleAssignPlatformRole}
              handleRemovePlatformRoleShortcut={
                handleRemovePlatformRoleShortcut
              }
              shortcutTenantId={shortcutTenantId}
              setShortcutTenantId={setShortcutTenantId}
              shortcutTenantUserId={shortcutTenantUserId}
              setShortcutTenantUserId={setShortcutTenantUserId}
              shortcutTenantRoleName={shortcutTenantRoleName}
              setShortcutTenantRoleName={setShortcutTenantRoleName}
              tenantOptions={tenantOptions}
              tenantRoleOptions={tenantRoleOptions}
              handleAssignTenantRole={handleAssignTenantRole}
              handleRemoveTenantMembershipShortcut={
                handleRemoveTenantMembershipShortcut
              }
            />
          </Show>

          <Show when={activeSection() === 'iam-principals'}>
            <AdminIamPrincipalProvider value={principalContextValue}>
              <PrincipalPoliciesPanel />
            </AdminIamPrincipalProvider>
          </Show>

          <Show when={activeSection() === 'iam-trust-sim'}>
            <AdminIamTrustSimProvider value={trustSimContextValue}>
              <TrustSimulationPanel />
            </AdminIamTrustSimProvider>
          </Show>
        </Card>
      </Show>
    </PageShell>
  )
}
