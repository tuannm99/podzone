export const roleOptions = [
  { name: 'Workspace owner', value: 'tenant_owner' },
  { name: 'Workspace admin', value: 'tenant_admin' },
  { name: 'Workspace operator', value: 'tenant_editor' },
  { name: 'Workspace viewer', value: 'tenant_viewer' },
]

export const platformRoleOptions = [
  { name: 'Platform owner', value: 'platform_owner' },
  { name: 'Platform admin', value: 'platform_admin' },
]

export function sessionStatusColor(status?: string) {
  return status === 'active' ? 'green' : 'dark'
}

export function membershipStatusColor(status?: string) {
  return status === 'active' ? 'green' : 'dark'
}

export function parseUserID(raw: unknown): number {
  if (typeof raw === 'number' && Number.isFinite(raw)) return raw
  if (typeof raw === 'string') {
    const parsed = Number.parseInt(raw, 10)
    return Number.isFinite(parsed) ? parsed : 0
  }
  return 0
}
