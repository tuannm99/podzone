import { createSignal } from 'solid-js'
import type { TenantMembership } from '@/services/iam'

export function createShellState() {
  const [error, setError] = createSignal('')
  const [message, setMessage] = createSignal('')
  const [loading, setLoading] = createSignal(false)
  const [allowed, setAllowed] = createSignal(false)
  const [memberships, setMemberships] = createSignal<TenantMembership[]>([])
  const tenantOptions = () =>
    memberships().map((membership) => ({
      name: `${membership.tenantId} · ${membership.roleName}`,
      value: membership.tenantId,
    }))

  return {
    pageError: error,
    setPageError: setError,
    pageMessage: message,
    setPageMessage: setMessage,
    loading,
    setLoading,
    allowed,
    setAllowed,
    memberships,
    setMemberships,
    tenantOptions,
  }
}
