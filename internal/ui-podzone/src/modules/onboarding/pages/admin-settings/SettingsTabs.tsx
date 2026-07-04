import { For } from 'solid-js'
import { classes } from '@/solid/shared/utils'
import type { AdminSettingsTab } from './context'

const tabs: Array<{
  value: AdminSettingsTab
  label: string
}> = [
  { value: 'overview', label: 'Overview' },
  { value: 'sessions', label: 'Sessions' },
  { value: 'team', label: 'Team access' },
  { value: 'invites', label: 'Invites' },
  { value: 'audit', label: 'Audit log' },
  { value: 'platform', label: 'Platform roles' },
]

export function SettingsTabs(props: {
  value: AdminSettingsTab
  onChange: (tab: AdminSettingsTab) => void
}) {
  return (
    <div
      role="tablist"
      aria-label="Settings sections"
      class="overflow-x-auto border-b border-gray-200"
    >
      <div class="flex min-w-max gap-6">
        <For each={tabs}>
          {(tab) => {
            const selected = () => props.value === tab.value
            return (
              <button
                type="button"
                role="tab"
                aria-selected={selected()}
                class={classes(
                  'border-b-2 px-1 py-3 text-sm font-medium transition',
                  selected()
                    ? 'border-gray-950 text-gray-950'
                    : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-900'
                )}
                onClick={() => props.onChange(tab.value)}
              >
                {tab.label}
              </button>
            )
          }}
        </For>
      </div>
    </div>
  )
}
