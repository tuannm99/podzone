import { Link, useRouterState } from '@tanstack/solid-router'
import { For, Show } from 'solid-js'
import { logout } from '@podzone/shared/services/auth'
import { tokenStorage } from '@podzone/shared/services/tokenStorage'
import { useTenantWorkspace } from '@podzone/shared/auth'
import { Button } from '@podzone/shared/ui/components/common/Primitives'
import { classes } from '@podzone/shared/ui/shared/utils'

type NavItem = {
    href: string
    label: string
    section: 'Platform' | 'Operations'
    active: boolean
}

function normalizePath(path: string) {
    if (path === '/') return path
    return path.replace(/\/+$/, '')
}

function initials(value: string) {
    const cleaned = value.trim()
    if (!cleaned) return 'PZ'
    return cleaned
        .split(/\s+/)
        .slice(0, 2)
        .map((part) => part[0]?.toUpperCase())
        .join('')
}

export function PodzoneNavbar() {
    const workspace = useTenantWorkspace()
    const user = tokenStorage.getUser()
    const pathname = useRouterState({
        select: (state) => normalizePath(state.location.pathname),
    })

    const hasTenant = () => (workspace?.tenantId() || tokenStorage.getActiveTenantID()).trim().length > 0
    const activeTenantId = () => workspace?.tenantId() || tokenStorage.getActiveTenantID()
    const currentStore = () => workspace?.currentStore()
    const currentStoreId = () => workspace?.currentStoreId() || ''
    const storeName = () => currentStore()?.name || currentStoreId() || 'Choose a store'
    const accountLabel = () => user?.username || user?.email || 'Account'
    const scopedHref = (path = '') => {
        const tenant = activeTenantId().trim()
        const storeId = currentStoreId().trim()
        const params = storeId ? `?storeId=${encodeURIComponent(storeId)}` : ''
        return `/t/${tenant}${path}${params}`
    }
    const isCurrent = (path: string, includeChildren = false) => {
        const currentPath = pathname()
        const targetPath = normalizePath(path)
        return currentPath === targetPath || (includeChildren && currentPath.startsWith(`${targetPath}/`))
    }

    const links = (): NavItem[] => [
        {
            href: '/admin',
            label: 'Home',
            section: 'Platform',
            active: isCurrent('/admin'),
        },
        {
            href: '/admin/provisioning',
            label: 'Provisioning',
            section: 'Platform',
            active: isCurrent('/admin/provisioning', true),
        },
        {
            href: '/admin/settings',
            label: 'Settings',
            section: 'Platform',
            active: isCurrent('/admin/settings', true),
        },
        {
            href: '/admin/iam',
            label: 'IAM',
            section: 'Platform',
            active: isCurrent('/admin/iam', true),
        },
        ...(hasTenant() && currentStoreId()
            ? [
                  {
                      href: scopedHref(),
                      label: 'Store Home',
                      section: 'Operations' as const,
                      active: isCurrent(`/t/${activeTenantId().trim()}`),
                  },
                  {
                      href: scopedHref('/orders'),
                      label: 'Orders',
                      section: 'Operations' as const,
                      active: isCurrent(`/t/${activeTenantId().trim()}/orders`),
                  },
                  {
                      href: scopedHref('/products/setup'),
                      label: 'Products',
                      section: 'Operations' as const,
                      active: isCurrent(`/t/${activeTenantId().trim()}/products/setup`, true),
                  },
                  {
                      href: scopedHref('/partners'),
                      label: 'Partners',
                      section: 'Operations' as const,
                      active: isCurrent(`/t/${activeTenantId().trim()}/partners`, true),
                  },
                  {
                      href: scopedHref('/orders/audit'),
                      label: 'Audit',
                      section: 'Operations' as const,
                      active: isCurrent(`/t/${activeTenantId().trim()}/orders/audit`, true),
                  },
                  {
                      href: scopedHref('/orders/finance'),
                      label: 'Finance',
                      section: 'Operations' as const,
                      active: isCurrent(`/t/${activeTenantId().trim()}/orders/finance`, true),
                  },
              ]
            : []),
    ]

    const platformLinks = () => links().filter((item) => item.section === 'Platform')
    const operationLinks = () => links().filter((item) => item.section === 'Operations')

    const navLinkClass = (active: boolean) =>
        classes(
            'flex h-10 items-center rounded-md px-3 text-sm font-medium transition',
            active ? 'bg-gray-900 text-white shadow-sm' : 'text-gray-600 hover:bg-gray-100 hover:text-gray-950'
        )

    return (
        <>
            <aside class="fixed inset-y-0 left-0 z-40 hidden w-64 border-r border-gray-200 bg-white lg:flex lg:flex-col">
                <div class="flex h-16 items-center gap-3 border-b border-gray-200 px-5">
                    <Link
                        to="/admin"
                        class="flex size-9 items-center justify-center rounded-md bg-gray-950 text-xs font-bold tracking-widest text-white"
                    >
                        PZ
                    </Link>
                    <div class="min-w-0">
                        <div class="truncate text-sm font-semibold text-gray-950">Podzone</div>
                        <div class="truncate text-xs text-gray-500">Seller Center</div>
                    </div>
                </div>

                <div class="border-b border-gray-200 p-4">
                    <Show
                        when={workspace}
                        fallback={
                            <div class="space-y-2">
                                <label class="text-xs font-semibold uppercase text-gray-500">Workspace</label>
                                <Button href="/admin" color="alternative" size="sm" class="w-full">
                                    Choose store
                                </Button>
                            </div>
                        }
                    >
                        <div class="space-y-2">
                            <label class="text-xs font-semibold uppercase text-gray-500">Current store</label>
                            <div class="truncate rounded-md border border-gray-200 bg-gray-50 px-3 py-2 text-sm font-medium text-gray-950">
                                {storeName()}
                            </div>
                            <div class="truncate text-xs text-gray-500">{activeTenantId().trim()}</div>
                            <Button href="/admin" color="alternative" size="sm" class="w-full">
                                Change store
                            </Button>
                        </div>
                    </Show>
                </div>

                <nav class="flex-1 space-y-6 overflow-y-auto p-4">
                    <div class="space-y-1">
                        <div class="px-3 text-xs font-semibold uppercase text-gray-400">Platform</div>
                        <For each={platformLinks()}>
                            {(link) => (
                                <Link
                                    to={link.href}
                                    class={navLinkClass(link.active)}
                                    aria-current={link.active ? 'page' : undefined}
                                >
                                    {link.label}
                                </Link>
                            )}
                        </For>
                    </div>

                    <Show when={operationLinks().length > 0}>
                        <div class="space-y-1">
                            <div class="px-3 text-xs font-semibold uppercase text-gray-400">Operations</div>
                            <For each={operationLinks()}>
                                {(link) => (
                                    <Link
                                        to={link.href}
                                        class={navLinkClass(link.active)}
                                        aria-current={link.active ? 'page' : undefined}
                                    >
                                        {link.label}
                                    </Link>
                                )}
                            </For>
                        </div>
                    </Show>
                </nav>

                <div class="border-t border-gray-200 p-4">
                    <div class="flex items-center gap-3">
                        <div class="flex size-9 shrink-0 items-center justify-center rounded-full bg-gray-100 text-xs font-semibold text-gray-700">
                            {initials(accountLabel())}
                        </div>
                        <div class="min-w-0 flex-1">
                            <div class="truncate text-sm font-medium text-gray-950">{accountLabel()}</div>
                            <div class="truncate text-xs text-gray-500">Signed in</div>
                        </div>
                    </div>
                    <Button
                        color="alternative"
                        size="sm"
                        class="mt-3 w-full"
                        onClick={() => {
                            void logout()
                        }}
                    >
                        Sign out
                    </Button>
                </div>
            </aside>

            <header class="sticky top-0 z-30 border-b border-gray-200 bg-white/95 backdrop-blur lg:pl-64">
                <div class="flex min-h-16 items-center justify-between gap-3 px-4 sm:px-6 lg:px-8">
                    <div class="min-w-0">
                        <div class="truncate text-sm font-semibold text-gray-950">{storeName()}</div>
                        <div class="truncate text-xs text-gray-500">
                            {hasTenant() ? activeTenantId().trim() : 'Platform workspace'}
                        </div>
                    </div>

                    <div class="flex min-w-0 flex-1 justify-end lg:hidden">
                        <nav class="flex gap-1 overflow-x-auto">
                            <For each={links()}>
                                {(link) => (
                                    <Link
                                        to={link.href}
                                        class={classes(
                                            'whitespace-nowrap rounded-md px-3 py-2 text-sm font-medium',
                                            link.active ? 'bg-gray-900 text-white' : 'text-gray-600 hover:bg-gray-100'
                                        )}
                                        aria-current={link.active ? 'page' : undefined}
                                    >
                                        {link.label}
                                    </Link>
                                )}
                            </For>
                        </nav>
                    </div>

                    <div class="hidden items-center gap-3 lg:flex">
                        <div class="text-right">
                            <div class="max-w-52 truncate text-sm font-medium text-gray-900">{accountLabel()}</div>
                            <div class="text-xs text-gray-500">Operator</div>
                        </div>
                        <div class="flex size-9 items-center justify-center rounded-full bg-gray-100 text-xs font-semibold text-gray-700">
                            {initials(accountLabel())}
                        </div>
                    </div>
                </div>
            </header>
        </>
    )
}
