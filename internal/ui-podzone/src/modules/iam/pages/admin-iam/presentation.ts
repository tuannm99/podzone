export const policyScopeOptions = [
  { name: 'Platform', value: 'platform' },
  { name: 'Tenant', value: 'tenant' },
]

export const groupScopeOptions = [
  { name: 'Platform', value: 'platform' },
  { name: 'Tenant', value: 'tenant' },
]

export const tenantRoleOptions = [
  { name: 'Workspace owner', value: 'tenant_owner' },
  { name: 'Workspace admin', value: 'tenant_admin' },
  { name: 'Workspace operator', value: 'tenant_editor' },
  { name: 'Workspace viewer', value: 'tenant_viewer' },
]

export const platformRoleOptions = [
  { name: 'Platform owner', value: 'platform_owner' },
  { name: 'Platform admin', value: 'platform_admin' },
]

export type IamSectionID =
  | 'iam-orgs'
  | 'iam-policies'
  | 'iam-groups'
  | 'iam-assignments'
  | 'iam-principals'
  | 'iam-trust-sim'

export type IamSection = {
  id: IamSectionID
  label: string
}

export const sectionLinks: IamSection[] = [
  { id: 'iam-orgs', label: 'Orgs & SCP' },
  { id: 'iam-policies', label: 'Policies' },
  { id: 'iam-groups', label: 'Groups' },
  { id: 'iam-assignments', label: 'Assignments' },
  { id: 'iam-principals', label: 'Principals' },
  { id: 'iam-trust-sim', label: 'Trust & simulator' },
]

export function prettyJSON(value: unknown) {
  return JSON.stringify(value, null, 2)
}

export function parseJSONArray<T>(raw: string, label: string) {
  const parsed = JSON.parse(raw || '[]')
  if (!Array.isArray(parsed)) {
    throw new Error(`${label} must be a JSON array`)
  }
  return parsed as T[]
}

export function parseJSONObject(raw: string, label: string) {
  const parsed = JSON.parse(raw || '{}')
  if (!parsed || Array.isArray(parsed) || typeof parsed !== 'object') {
    throw new Error(`${label} must be a JSON object`)
  }
  return parsed as Record<string, string>
}

export function attachmentColor(type: string) {
  if (type.includes('boundary')) return 'pink'
  if (type.includes('service_control')) return 'yellow'
  if (type.includes('group')) return 'green'
  return 'blue'
}

export function simulationSourceColor(source: string) {
  const normalized = source.toLowerCase()
  if (normalized.includes('deny')) return 'red'
  if (normalized.includes('boundary')) return 'pink'
  if (normalized.includes('scp')) return 'yellow'
  if (normalized.includes('session')) return 'indigo'
  if (normalized.includes('group')) return 'green'
  return 'blue'
}

export function simulationLayerTone(allowed: boolean, reason: string) {
  if (!allowed) {
    if (reason.toLowerCase().includes('deny')) {
      return 'border-red-200 bg-red-50'
    }
    return 'border-amber-200 bg-amber-50'
  }
  return 'border-green-200 bg-green-50'
}

export function statementSourceLabel(source: string) {
  return source.replaceAll('_', ' ')
}
