import { For, Show } from 'solid-js'
import {
  EmptyBlock,
  InfoAlert,
  LoadingInline,
} from '@/solid/components/common/Feedback'
import { Badge, Button } from '@/solid/components/common/Primitives'
import { useAdminSettings } from '../context'
import { membershipStatusColor } from '../presentation'

export function TeamMembersList() {
  const { teamAccess } = useAdminSettings()
  const { access } = teamAccess

  return (
    <>
      <Show when={access.loadingAccess()}>
        <LoadingInline label="Loading workspace team..." />
      </Show>
      <Show
        when={!access.loadingAccess() && access.members().length > 0}
        fallback={
          <EmptyBlock
            title="No team members loaded"
            copy="Choose a workspace and reload team access to inspect who can operate in that workspace."
          />
        }
      >
        <Show when={!access.canRead()}>
          <EmptyBlock
            title="No workspace access"
            copy="You do not currently have permission to inspect team access for this workspace."
          />
        </Show>
        <Show when={access.canRead() && !access.canManage()}>
          <InfoAlert>
            You can inspect this workspace, but only authorized workspace owners
            or admins can manage team access.
          </InfoAlert>
        </Show>
        <div class="space-y-3">
          <For each={access.members()}>
            {(membership) => (
              <div class="rounded-lg border border-gray-200 p-4">
                <div class="flex flex-wrap items-center justify-between gap-3">
                  <div>
                    <p class="font-semibold text-gray-900">
                      teammate {membership.userId}
                    </p>
                    <p class="mt-1 text-sm text-gray-500">
                      {membership.tenantId}
                    </p>
                  </div>
                  <div class="flex flex-wrap items-center gap-2">
                    <Badge content={membership.roleName} color="blue" />
                    <Badge
                      content={membership.status}
                      color={membershipStatusColor(membership.status)}
                    />
                    <Button
                      color="red"
                      size="xs"
                      disabled={!access.canManage()}
                      onClick={() =>
                        void teamAccess.remove(
                          membership.tenantId,
                          membership.userId
                        )
                      }
                    >
                      Remove
                    </Button>
                  </div>
                </div>
              </div>
            )}
          </For>
        </div>
      </Show>
    </>
  )
}
