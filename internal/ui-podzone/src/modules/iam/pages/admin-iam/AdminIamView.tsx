import { Show, createEffect, createSignal, onMount } from 'solid-js'
import type { AdminIamViewModel } from './createAdminIamViewModel'
import {
  EmptyBlock,
  ErrorAlert,
  InfoAlert,
  LoadingInline,
} from '@/solid/components/common/Feedback'
import { PageShell } from '@/solid/components/common/PageShell'
import { Button, Card } from '@/solid/components/common/Primitives'
import { SectionLead } from '@/solid/components/common/SectionLead'
import { RoleAssignmentsPanel } from './assignments/RoleAssignmentsPanel'
import { AdminIamGroupProvider } from './groups/context'
import { GroupsPanel } from './groups/GroupsPanel'
import { IamWorkspaceNav } from './IamWorkspaceNav'
import { OrganizationsPanel } from './organizations/OrganizationsPanel'
import { AdminIamPolicyProvider } from './policies/context'
import { PoliciesPanel } from './policies/PoliciesPanel'
import { type IamSection, type IamSectionID } from './presentation'
import { AdminIamPrincipalProvider } from './principals/context'
import { PrincipalPoliciesPanel } from './principals/PrincipalPoliciesPanel'
import { AdminIamTrustSimProvider } from './trust-simulation/context'
import { TrustSimulationPanel } from './trust-simulation/TrustSimulationPanel'

// The controller page assembles the model; this view only owns composition.
export function AdminIamView(props: { model: AdminIamViewModel }) {
  const sections = () => {
    const allSections = props.model.sectionLinks as IamSection[]
    if (props.model.feedback.canManagePlatform()) return allSections
    return allSections.filter(
      (section) => section.id === 'iam-orgs' || section.id === 'iam-assignments'
    )
  }
  const [activeSection, setActiveSection] =
    createSignal<IamSectionID>('iam-orgs')

  const selectSection = (section: IamSectionID) => {
    setActiveSection(section)
    window.history.replaceState(null, '', `#${section}`)
  }

  onMount(() => {
    const hash = window.location.hash.slice(1)
    const selected = sections().find((section) => section.id === hash)
    if (selected) setActiveSection(selected.id)
  })

  createEffect(() => {
    if (!sections().some((section) => section.id === activeSection())) {
      setActiveSection('iam-orgs')
    }
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

      <Show when={props.model.feedback.error()}>
        <ErrorAlert>{props.model.feedback.error()}</ErrorAlert>
      </Show>
      <Show when={props.model.feedback.message()}>
        <InfoAlert>{props.model.feedback.message()}</InfoAlert>
      </Show>
      <Show when={props.model.feedback.loading()}>
        <LoadingInline label="Loading IAM control plane..." />
      </Show>
      <Show
        when={
          !props.model.feedback.loading() && !props.model.feedback.allowed()
        }
      >
        <EmptyBlock
          title="IAM console unavailable"
          copy="This session does not have the required platform permission."
        />
      </Show>

      <Show when={props.model.feedback.allowed()}>
        <IamWorkspaceNav
          sections={sections()}
          activeSection={activeSection()}
          onSelect={selectSection}
        />

        <Card class="space-y-4">
          <Show when={activeSection() === 'iam-orgs'}>
            <OrganizationsPanel
              organizationOptions={
                props.model.organizations.organizationOptions
              }
              selectedOrgId={props.model.organizations.selectedOrgId}
              setSelectedOrgId={props.model.organizations.setSelectedOrgId}
              submitCreateOrganization={
                props.model.organizations.submitCreateOrganization
              }
              orgName={props.model.organizations.orgName}
              setOrgName={props.model.organizations.setOrgName}
              orgSlug={props.model.organizations.orgSlug}
              setOrgSlug={props.model.organizations.setOrgSlug}
              orgTenantId={props.model.organizations.orgTenantId}
              setOrgTenantId={props.model.organizations.setOrgTenantId}
              orgPolicyName={props.model.organizations.orgPolicyName}
              setOrgPolicyName={props.model.organizations.setOrgPolicyName}
              tenantOptions={props.model.organizations.tenantOptions}
              handleAttachTenantToOrg={
                props.model.organizations.handleAttachTenantToOrg
              }
              handleAttachScp={props.model.organizations.handleAttachScp}
              organizations={props.model.organizations.items}
              query={props.model.organizations.query}
              pageInfo={props.model.organizations.pageInfo}
              loading={props.model.organizations.loading}
              error={props.model.organizations.error}
              updateQuery={props.model.organizations.updateQuery}
              orgPolicies={props.model.organizations.orgPolicies}
              handleDetachTenantFromOrg={
                props.model.organizations.handleDetachTenantFromOrg
              }
              handleDetachScp={props.model.organizations.handleDetachScp}
              canManagePlatform={props.model.organizations.canManagePlatform}
              organizationMembers={
                props.model.organizations.organizationMembers
              }
              organizationMembersQuery={
                props.model.organizations.organizationMembersQuery
              }
              organizationMembersPageInfo={
                props.model.organizations.organizationMembersPageInfo
              }
              organizationMembersLoading={
                props.model.organizations.organizationMembersLoading
              }
              organizationMembersError={
                props.model.organizations.organizationMembersError
              }
              updateOrganizationMembersQuery={
                props.model.organizations.updateOrganizationMembersQuery
              }
              handleAddOrganizationMember={
                props.model.organizations.handleAddOrganizationMember
              }
              handleRemoveOrganizationMember={
                props.model.organizations.handleRemoveOrganizationMember
              }
            />
          </Show>

          <Show when={activeSection() === 'iam-policies'}>
            <AdminIamPolicyProvider value={props.model.policies}>
              <PoliciesPanel />
            </AdminIamPolicyProvider>
          </Show>

          <Show when={activeSection() === 'iam-groups'}>
            <AdminIamGroupProvider value={props.model.groups}>
              <GroupsPanel />
            </AdminIamGroupProvider>
          </Show>

          <Show when={activeSection() === 'iam-assignments'}>
            <RoleAssignmentsPanel
              shortcutPlatformUserId={props.model.assignments.platformUserId}
              setShortcutPlatformUserId={
                props.model.assignments.setPlatformUserId
              }
              shortcutPlatformRoleName={
                props.model.assignments.platformRoleName
              }
              setShortcutPlatformRoleName={
                props.model.assignments.setPlatformRoleName
              }
              platformRoleOptions={props.model.assignments.platformRoleOptions}
              handleAssignPlatformRole={
                props.model.assignments.assignPlatformRole
              }
              handleRemovePlatformRoleShortcut={
                props.model.assignments.removePlatformRole
              }
              shortcutTenantId={props.model.assignments.tenantId}
              setShortcutTenantId={props.model.assignments.setTenantId}
              shortcutTenantUserId={props.model.assignments.tenantUserId}
              setShortcutTenantUserId={props.model.assignments.setTenantUserId}
              shortcutTenantRoleName={props.model.assignments.tenantRoleName}
              setShortcutTenantRoleName={
                props.model.assignments.setTenantRoleName
              }
              tenantOptions={props.model.assignments.tenantOptions}
              tenantRoleOptions={props.model.assignments.tenantRoleOptions}
              handleAssignTenantRole={props.model.assignments.assignTenantRole}
              handleRemoveTenantMembershipShortcut={
                props.model.assignments.removeTenantMembership
              }
              canManagePlatform={props.model.assignments.canManagePlatform}
            />
          </Show>

          <Show when={activeSection() === 'iam-principals'}>
            <AdminIamPrincipalProvider value={props.model.principals}>
              <PrincipalPoliciesPanel />
            </AdminIamPrincipalProvider>
          </Show>

          <Show when={activeSection() === 'iam-trust-sim'}>
            <AdminIamTrustSimProvider value={props.model.trustSimulation}>
              <TrustSimulationPanel />
            </AdminIamTrustSimProvider>
          </Show>
        </Card>
      </Show>
    </PageShell>
  )
}
