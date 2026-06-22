import { For, Show } from 'solid-js';
import {
  EmptyBlock,
  ErrorAlert,
  InfoAlert,
  LoadingInline,
} from '@/solid/components/common/Feedback';
import { PageShell } from '@/solid/components/common/PageShell';
import {
  Badge,
  Button,
  Card,
  InputField,
  SelectField,
} from '@/solid/components/common/Primitives';
import { SectionLead } from '@/solid/components/common/SectionLead';
import { SectionTitle } from '@/solid/components/common/SectionTitle';
import { AdminIamGroupProvider } from './group-context';
import { GroupsPanel } from './GroupsPanel';
import { AdminIamPolicyProvider } from './policy-context';
import { PoliciesPanel } from './PoliciesPanel';
import { AdminIamPrincipalProvider } from './principal-context';
import { PrincipalPoliciesPanel } from './PrincipalPoliciesPanel';
import { AdminIamTrustSimProvider } from './trust-sim-context';
import { TrustSimulationPanel } from './TrustSimulationPanel';

// The IAM page model is intentionally assembled by the controller page while
// this component only owns layout composition.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function AdminIamView(props: { model: any }) {
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
  } = props.model;

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
        <Card class="space-y-3">
          <SectionTitle
            title="Jump to section"
            subtitle="Use this instead of scrolling through the entire IAM console."
          />
          <div class="flex flex-wrap gap-2">
            <For each={sectionLinks}>
              {(section) => (
                <Button href={`#${section.id}`} size="sm" color="alternative">
                  {section.label}
                </Button>
              )}
            </For>
          </div>
        </Card>

        <div class="grid gap-6 xl:grid-cols-2">
          <div id="iam-orgs" class="scroll-mt-24">
            <Card class="space-y-4">
              <SectionTitle
                title="Organizations and SCP"
                subtitle="Create organizations, map workspaces, and attach service control policies."
              />
              <form
                class="grid gap-3 md:grid-cols-2"
                onSubmit={submitCreateOrganization}
              >
                <InputField
                  label="Organization name"
                  value={orgName()}
                  onInput={(e) => setOrgName(e.currentTarget.value)}
                />
                <InputField
                  label="Organization slug"
                  value={orgSlug()}
                  onInput={(e) => setOrgSlug(e.currentTarget.value)}
                />
                <div class="md:col-span-2">
                  <Button type="submit" size="sm">
                    Create organization
                  </Button>
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
                <Button
                  size="sm"
                  onClick={handleAttachTenantToOrg}
                  disabled={!selectedOrgId() || !orgTenantId()}
                >
                  Attach workspace
                </Button>
                <Button
                  size="sm"
                  color="dark"
                  onClick={handleAttachScp}
                  disabled={!selectedOrgId() || !orgPolicyName().trim()}
                >
                  Attach SCP
                </Button>
              </div>

              <Show
                when={organizations().length > 0}
                fallback={
                  <EmptyBlock
                    title="No organizations"
                    copy="Create the first organization to start applying SCP guardrails."
                  />
                }
              >
                <div class="space-y-3">
                  <For each={organizations()}>
                    {(org) => (
                      <div class="rounded-lg border border-gray-200 p-4">
                        <div class="flex flex-wrap items-center justify-between gap-3">
                          <div>
                            <p class="font-semibold text-gray-900">
                              {org.name}
                            </p>
                            <p class="text-sm text-gray-500">{org.slug}</p>
                          </div>
                          <Badge
                            content={
                              org.id === selectedOrgId()
                                ? 'selected'
                                : 'organization'
                            }
                            color={org.id === selectedOrgId() ? 'blue' : 'dark'}
                          />
                        </div>
                        <div class="mt-3 flex flex-wrap gap-2">
                          <Show
                            when={
                              org.id === selectedOrgId() &&
                              orgTenantId().trim()
                            }
                          >
                            <Button
                              size="xs"
                              color="light"
                              onClick={() =>
                                handleDetachTenantFromOrg(orgTenantId().trim())
                              }
                            >
                              Detach selected workspace
                            </Button>
                          </Show>
                          <For
                            each={org.id === selectedOrgId() ? orgPolicies() : []}
                          >
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
          </div>

          <div id="iam-policies" class="scroll-mt-24">
            <Card class="space-y-4">
              <AdminIamPolicyProvider value={policyContextValue}>
                <PoliciesPanel />
              </AdminIamPolicyProvider>
            </Card>
          </div>

          <div id="iam-groups" class="scroll-mt-24">
            <Card class="space-y-4">
              <AdminIamGroupProvider value={groupContextValue}>
                <GroupsPanel />
              </AdminIamGroupProvider>
            </Card>
          </div>

          <div id="iam-shortcuts" class="scroll-mt-24">
            <Card class="space-y-4">
              <SectionTitle
                title="Role assignment shortcuts"
                subtitle="Quickly grant or revoke platform roles and workspace memberships without leaving the IAM console."
              />
              <div class="grid gap-6 lg:grid-cols-2">
                <div class="space-y-3 rounded-lg border border-gray-200 p-4">
                  <p class="text-sm font-semibold text-gray-900">
                    Platform role shortcut
                  </p>
                  <InputField
                    label="Target user id"
                    value={shortcutPlatformUserId()}
                    onInput={(e) =>
                      setShortcutPlatformUserId(e.currentTarget.value)
                    }
                  />
                  <SelectField
                    label="Platform role"
                    value={shortcutPlatformRoleName()}
                    options={platformRoleOptions}
                    onChange={(e) =>
                      setShortcutPlatformRoleName(e.currentTarget.value)
                    }
                  />
                  <div class="flex flex-wrap gap-3">
                    <Button
                      size="sm"
                      onClick={handleAssignPlatformRole}
                      disabled={!shortcutPlatformUserId().trim()}
                    >
                      Assign platform role
                    </Button>
                    <Button
                      size="sm"
                      color="red"
                      onClick={handleRemovePlatformRoleShortcut}
                      disabled={!shortcutPlatformUserId().trim()}
                    >
                      Remove platform role
                    </Button>
                  </div>
                </div>

                <div class="space-y-3 rounded-lg border border-gray-200 p-4">
                  <p class="text-sm font-semibold text-gray-900">
                    Workspace membership shortcut
                  </p>
                  <SelectField
                    label="Workspace"
                    value={shortcutTenantId()}
                    options={tenantOptions()}
                    onChange={(e) => setShortcutTenantId(e.currentTarget.value)}
                  />
                  <InputField
                    label="Target user id"
                    value={shortcutTenantUserId()}
                    onInput={(e) =>
                      setShortcutTenantUserId(e.currentTarget.value)
                    }
                  />
                  <SelectField
                    label="Workspace role"
                    value={shortcutTenantRoleName()}
                    options={tenantRoleOptions}
                    onChange={(e) =>
                      setShortcutTenantRoleName(e.currentTarget.value)
                    }
                  />
                  <div class="flex flex-wrap gap-3">
                    <Button
                      size="sm"
                      onClick={handleAssignTenantRole}
                      disabled={
                        !shortcutTenantId().trim() ||
                        !shortcutTenantUserId().trim()
                      }
                    >
                      Assign workspace role
                    </Button>
                    <Button
                      size="sm"
                      color="red"
                      onClick={handleRemoveTenantMembershipShortcut}
                      disabled={
                        !shortcutTenantId().trim() ||
                        !shortcutTenantUserId().trim()
                      }
                    >
                      Remove membership
                    </Button>
                  </div>
                </div>
              </div>
            </Card>
          </div>

          <div id="iam-principals" class="scroll-mt-24">
            <Card class="space-y-4">
              <AdminIamPrincipalProvider value={principalContextValue}>
                <PrincipalPoliciesPanel />
              </AdminIamPrincipalProvider>
            </Card>
          </div>

          <div id="iam-trust-sim" class="scroll-mt-24">
            <Card class="space-y-4">
              <AdminIamTrustSimProvider value={trustSimContextValue}>
                <TrustSimulationPanel />
              </AdminIamTrustSimProvider>
            </Card>
          </div>
        </div>
      </Show>
    </PageShell>
  );
}
