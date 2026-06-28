import { For } from 'solid-js'
import type { IamSection, IamSectionID } from './presentation'

type IamWorkspaceNavProps = {
  sections: IamSection[]
  activeSection: IamSectionID
  onSelect: (section: IamSectionID) => void
}

export function IamWorkspaceNav(props: IamWorkspaceNavProps) {
  return (
    <nav
      class="overflow-x-auto border-b border-gray-200"
      aria-label="IAM sections"
    >
      <div class="flex min-w-max gap-1">
        <For each={props.sections}>
          {(section) => (
            <button
              type="button"
              class={`border-b-2 px-4 py-3 text-sm font-medium transition ${
                props.activeSection === section.id
                  ? 'border-gray-950 text-gray-950'
                  : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-800'
              }`}
              aria-current={
                props.activeSection === section.id ? 'page' : undefined
              }
              onClick={() => props.onSelect(section.id)}
            >
              {section.label}
            </button>
          )}
        </For>
      </div>
    </nav>
  )
}
