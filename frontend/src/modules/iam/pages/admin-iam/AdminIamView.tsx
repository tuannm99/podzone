import { useNavigate, useSearch } from '@tanstack/solid-router'
import { Show } from 'solid-js'
import type { AdminIamViewModel } from './createAdminIamViewModel'
import { EmptyBlock, ErrorAlert, InfoAlert, LoadingInline } from '@/solid/components/common/Feedback'
import { PageShell } from '@/solid/components/common/PageShell'
import { Button } from '@/solid/components/common/Primitives'
import { Tabs } from '@/solid/components/common/Tabs'
import { RoleAssignmentsPanel } from './assignments/RoleAssignmentsPanel'
import { AdminIamGroupProvider } from './groups/context'
import { GroupsPanel } from './groups/GroupsPanel'
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
            (section) =>
                section.id === 'iam-orgs' ||
                section.id === 'iam-policies' ||
                section.id === 'iam-groups' ||
                section.id === 'iam-assignments'
        )
    }
    const sectionTabs = () => sections().map((section) => ({ value: section.id, label: section.label }))
    const navigate = useNavigate()
    const search = useSearch({ from: '/admin/iam' })
    const activeSection = (): IamSectionID => {
        const s = search().section
        return sections().some((sec) => sec.id === s) ? (s as IamSectionID) : 'iam-orgs'
    }

    const selectSection = (section: IamSectionID) => {
        void navigate({ to: '/admin/iam', search: (prev) => ({ ...prev, section }) })
    }

    return (
        <PageShell>
            <header class="flex flex-col gap-4 border-b border-gray-200 pb-5 lg:flex-row lg:items-end lg:justify-between">
                <div class="space-y-2">
                    <div class="text-xs font-semibold uppercase text-gray-500">IAM Console</div>
                    <h1 class="text-2xl font-semibold text-gray-950">Identity and access control</h1>
                    <p class="text-sm text-gray-600">Organizations, policies, principals, and access evaluation.</p>
                </div>
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
            <Show when={!props.model.feedback.loading() && !props.model.feedback.allowed()}>
                <EmptyBlock
                    title="IAM console unavailable"
                    copy="This session does not have the required platform permission."
                />
            </Show>

            <Show when={props.model.feedback.allowed()}>
                <Tabs
                    ariaLabel="IAM sections"
                    items={sectionTabs()}
                    value={activeSection()}
                    onChange={selectSection}
                    variant="underline"
                />

                <div role="tabpanel" class="min-w-0 pt-1">
                    <Show when={activeSection() === 'iam-orgs'}>
                        <OrganizationsPanel
                            organizationOptions={props.model.organizations.organizationOptions}
                            identityForUser={props.model.organizations.identityForUser}
                            policyOptions={props.model.organizations.policyOptions}
                            userOptions={props.model.organizations.userOptions}
                            usersLoading={props.model.organizations.usersLoading}
                            usersError={props.model.organizations.usersError}
                            searchUsers={props.model.organizations.searchUsers}
                            selectedOrgId={props.model.organizations.selectedOrgId}
                            setSelectedOrgId={props.model.organizations.setSelectedOrgId}
                            submitCreateOrganization={props.model.organizations.submitCreateOrganization}
                            orgName={props.model.organizations.orgName}
                            setOrgName={props.model.organizations.setOrgName}
                            orgSlug={props.model.organizations.orgSlug}
                            setOrgSlug={props.model.organizations.setOrgSlug}
                            orgTenantId={props.model.organizations.orgTenantId}
                            setOrgTenantId={props.model.organizations.setOrgTenantId}
                            orgPolicyName={props.model.organizations.orgPolicyName}
                            setOrgPolicyName={props.model.organizations.setOrgPolicyName}
                            tenantOptions={props.model.organizations.tenantOptions}
                            handleAttachTenantToOrg={props.model.organizations.handleAttachTenantToOrg}
                            handleAttachScp={props.model.organizations.handleAttachScp}
                            organizations={props.model.organizations.items}
                            query={props.model.organizations.query}
                            pageInfo={props.model.organizations.pageInfo}
                            loading={props.model.organizations.loading}
                            error={props.model.organizations.error}
                            updateQuery={props.model.organizations.updateQuery}
                            orgPolicies={props.model.organizations.orgPolicies}
                            handleDetachTenantFromOrg={props.model.organizations.handleDetachTenantFromOrg}
                            handleDetachScp={props.model.organizations.handleDetachScp}
                            canManagePlatform={props.model.organizations.canManagePlatform}
                            organizationMembers={props.model.organizations.organizationMembers}
                            organizationMembersQuery={props.model.organizations.organizationMembersQuery}
                            organizationMembersPageInfo={props.model.organizations.organizationMembersPageInfo}
                            organizationMembersLoading={props.model.organizations.organizationMembersLoading}
                            organizationMembersError={props.model.organizations.organizationMembersError}
                            updateOrganizationMembersQuery={props.model.organizations.updateOrganizationMembersQuery}
                            handleAddOrganizationMember={props.model.organizations.handleAddOrganizationMember}
                            handleRemoveOrganizationMember={props.model.organizations.handleRemoveOrganizationMember}
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
                            setShortcutPlatformUserId={props.model.assignments.setPlatformUserId}
                            platformUserOptions={props.model.assignments.platformUserOptions}
                            platformUsersLoading={props.model.assignments.platformUsersLoading}
                            platformUsersError={props.model.assignments.platformUsersError}
                            searchPlatformUsers={props.model.assignments.searchPlatformUsers}
                            shortcutPlatformRoleName={props.model.assignments.platformRoleName}
                            setShortcutPlatformRoleName={props.model.assignments.setPlatformRoleName}
                            platformRoleOptions={props.model.assignments.platformRoleOptions}
                            handleAssignPlatformRole={props.model.assignments.assignPlatformRole}
                            handleRemovePlatformRoleShortcut={props.model.assignments.removePlatformRole}
                            shortcutTenantId={props.model.assignments.tenantId}
                            setShortcutTenantId={props.model.assignments.setTenantId}
                            shortcutTenantUserId={props.model.assignments.tenantUserId}
                            setShortcutTenantUserId={props.model.assignments.setTenantUserId}
                            tenantUserOptions={props.model.assignments.tenantUserOptions}
                            tenantUsersLoading={props.model.assignments.tenantUsersLoading}
                            tenantUsersError={props.model.assignments.tenantUsersError}
                            searchTenantUsers={props.model.assignments.searchTenantUsers}
                            shortcutTenantRoleName={props.model.assignments.tenantRoleName}
                            setShortcutTenantRoleName={props.model.assignments.setTenantRoleName}
                            tenantOptions={props.model.assignments.tenantOptions}
                            tenantRoleOptions={props.model.assignments.tenantRoleOptions}
                            handleAssignTenantRole={props.model.assignments.assignTenantRole}
                            handleRemoveTenantMembershipShortcut={props.model.assignments.removeTenantMembership}
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
                </div>
            </Show>
        </PageShell>
    )
}
