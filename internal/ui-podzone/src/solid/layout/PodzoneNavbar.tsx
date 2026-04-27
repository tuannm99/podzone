import { Link, useNavigate } from '@tanstack/solid-router'
import { Show, createSignal } from 'solid-js'
import { logout } from '../../services/auth'
import { tenantStorage } from '../../services/tenantStorage'
import { tokenStorage } from '../../services/tokenStorage'
import { Badge, Button } from '../components/common/Primitives'
import { classes } from '../shared/utils'

export function PodzoneNavbar(props: { currentPath: string }) {
  const navigate = useNavigate()
  const [tenantId, setTenantId] = createSignal(tenantStorage.getTenantID())

  const links = [
    {
      href: '/admin',
      label: 'Overview',
      active: () => props.currentPath === '/admin'
    },
    {
      href: '/admin/settings',
      label: 'Settings',
      active: () => props.currentPath === '/admin/settings'
    }
  ]

  const user = tokenStorage.getUser()
  const hasTenant = () => tenantId().trim().length > 0

  const goToTenant = () => {
    const nextTenant = tenantId().trim()
    if (!nextTenant) return
    tenantStorage.setTenantID(nextTenant)
    void navigate({ to: `/t/${nextTenant}` })
  }

  return (
    <header class="-mx-4 sticky top-0 z-40 border-b border-gray-200 bg-white/85 shadow-sm backdrop-blur sm:-mx-6 lg:-mx-8">
      <div class="flex min-h-16 flex-wrap items-center justify-between gap-3 px-4 py-3 sm:px-6 lg:px-8">
        <Link to="/admin" class="flex min-w-0 items-center gap-3">
          <span class="inline-flex size-10 shrink-0 items-center justify-center rounded-xl bg-blue-600 text-xs font-bold uppercase tracking-[0.22em] text-white">
            pz
          </span>
          <div class="min-w-0">
            <div class="truncate text-sm font-semibold text-gray-900">Podzone Console</div>
            <div class="hidden text-xs text-gray-500 sm:block">Admin and tenant workspace</div>
          </div>
        </Link>

        <nav class="flex items-center gap-1">
          {links.map((link) => (
            <Link
              to={link.href}
              class={classes(
                'rounded-full px-3 py-1.5 text-sm font-medium transition',
                link.active()
                  ? 'bg-blue-50 text-blue-700'
                  : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'
              )}
            >
              {link.label}
            </Link>
          ))}
        </nav>

        <div class="flex flex-wrap items-center justify-end gap-2">
          <div class="flex items-center gap-2 rounded-full border border-gray-200 bg-gray-50 px-2 py-1.5">
            <input
              class="w-36 border-0 bg-transparent px-2 text-sm text-gray-700 outline-none placeholder:text-gray-400"
              value={tenantId()}
              placeholder="tenant id"
              onInput={(event) => setTenantId(event.currentTarget.value)}
            />
            <Button
              color="blue"
              size="xs"
              pill
              disabled={!hasTenant()}
              onClick={goToTenant}
            >
              Open
            </Button>
          </div>

          <Show when={hasTenant()}>
            <Badge content={`tenant ${tenantId().trim()}`} color="indigo" />
          </Show>

          <Badge content={user?.username || user?.email || 'authenticated'} color="dark" />

          <Button color="alternative" size="sm" pill onClick={() => logout()}>
            Sign out
          </Button>
        </div>
      </div>
    </header>
  )
}
