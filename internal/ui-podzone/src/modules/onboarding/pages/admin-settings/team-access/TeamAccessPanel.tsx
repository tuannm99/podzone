import { MyWorkspaceAccess } from './MyWorkspaceAccess'
import { TeamAccessEditor } from './TeamAccessEditor'

export function TeamAccessPanel() {
  return (
    <div class="grid gap-6 lg:grid-cols-[0.95fr_1.05fr]">
      <MyWorkspaceAccess />
      <TeamAccessEditor />
    </div>
  )
}
