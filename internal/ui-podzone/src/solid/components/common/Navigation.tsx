import { For, Show, type JSX, type ParentProps } from 'solid-js'
import { classes } from '../../shared/utils'

export type NavItem = {
  label: string
  href?: string
  active?: boolean
  icon?: JSX.Element
  onClick?: () => void
}

function NavAction(props: { item: NavItem; class?: string; activeClass?: string }) {
  const className = () =>
    classes(
      'inline-flex items-center gap-2 rounded-xl px-3 py-2 text-sm font-medium transition',
      props.item.active
        ? (props.activeClass ?? 'bg-blue-700 text-white')
        : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900',
      props.class
    )

  const content = (
    <>
      <Show when={props.item.icon}>
        <span class="shrink-0">{props.item.icon}</span>
      </Show>
      <span>{props.item.label}</span>
    </>
  )

  return props.item.href ? (
    <a href={props.item.href} class={className()}>
      {content}
    </a>
  ) : (
    <button type="button" class={className()} onClick={() => props.item.onClick?.()}>
      {content}
    </button>
  )
}

export function Navbar(props: {
  brand: JSX.Element
  items?: NavItem[]
  actions?: JSX.Element
  class?: string
}) {
  return (
    <header
      class={classes(
        'rounded-2xl border border-gray-200 bg-white px-4 py-3 shadow-sm',
        props.class
      )}
    >
      <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
        <div class="flex items-center justify-between gap-4">
          <div>{props.brand}</div>
          <Show when={props.actions}>
            <div class="lg:hidden">{props.actions}</div>
          </Show>
        </div>
        <div class="flex flex-1 flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
          <nav class="flex flex-wrap items-center gap-2">
            <For each={props.items ?? []}>{(item) => <NavAction item={item} />}</For>
          </nav>
          <Show when={props.actions}>
            <div class="hidden lg:block">{props.actions}</div>
          </Show>
        </div>
      </div>
    </header>
  )
}

export function Sidebar(
  props: ParentProps<{
    title?: string
    items: NavItem[]
    footer?: JSX.Element
    class?: string
  }>
) {
  return (
    <aside
      class={classes('rounded-2xl border border-gray-200 bg-white p-4 shadow-sm', props.class)}
    >
      <div class="space-y-4">
        <Show when={props.title || props.children}>
          <div class="space-y-2">
            <Show when={props.title}>
              <p class="text-xs font-semibold uppercase tracking-[0.22em] text-gray-400">
                {props.title}
              </p>
            </Show>
            {props.children}
          </div>
        </Show>
        <nav class="space-y-1">
          <For each={props.items}>
            {(item) => (
              <NavAction
                item={item}
                class="flex w-full justify-start"
                activeClass="bg-blue-50 text-blue-900"
              />
            )}
          </For>
        </nav>
        <Show when={props.footer}>
          <div class="border-t border-gray-100 pt-4">{props.footer}</div>
        </Show>
      </div>
    </aside>
  )
}

export function BottomNavigation(props: { items: NavItem[]; class?: string }) {
  return (
    <nav
      class={classes(
        'fixed inset-x-0 bottom-0 z-40 border-t border-gray-200 bg-white/95 px-4 py-2 shadow-[0_-8px_24px_rgba(15,23,42,0.08)] backdrop-blur md:hidden',
        props.class
      )}
    >
      <div class="mx-auto flex max-w-xl items-center justify-around gap-2">
        <For each={props.items}>
          {(item) => (
            <NavAction
              item={item}
              class="min-w-0 flex-1 flex-col justify-center gap-1 px-2 py-1.5 text-xs"
              activeClass="bg-transparent text-blue-700"
            />
          )}
        </For>
      </div>
    </nav>
  )
}

export function Footer(
  props: ParentProps<{
    brand?: JSX.Element
    links?: NavItem[]
    note?: string
    class?: string
  }>
) {
  return (
    <footer
      class={classes(
        'rounded-2xl border border-gray-200 bg-white px-6 py-5 shadow-sm',
        props.class
      )}
    >
      <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div class="space-y-2">
          <Show when={props.brand}>
            <div>{props.brand}</div>
          </Show>
          <Show when={props.note}>
            <p class="text-sm text-gray-500">{props.note}</p>
          </Show>
          <Show when={props.children}>
            <div>{props.children}</div>
          </Show>
        </div>
        <Show when={(props.links?.length ?? 0) > 0}>
          <nav class="flex flex-wrap items-center gap-3 text-sm text-gray-500">
            <For each={props.links ?? []}>
              {(item) => (
                <NavAction
                  item={item}
                  class="px-0 py-0 hover:bg-transparent"
                  activeClass="bg-transparent text-blue-700"
                />
              )}
            </For>
          </nav>
        </Show>
      </div>
    </footer>
  )
}

export type SpeedDialItem = {
  label: string
  href?: string
  onClick?: () => void
}

export function SpeedDial(props: { items: SpeedDialItem[]; class?: string }) {
  return (
    <div class={classes('fixed bottom-6 right-6 z-40 hidden flex-col gap-2 md:flex', props.class)}>
      <For each={props.items}>
        {(item) =>
          item.href ? (
            <a
              href={item.href}
              class="rounded-full bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-lg ring-1 ring-gray-200 transition hover:bg-gray-50"
            >
              {item.label}
            </a>
          ) : (
            <button
              type="button"
              class="rounded-full bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-lg ring-1 ring-gray-200 transition hover:bg-gray-50"
              onClick={() => item.onClick?.()}
            >
              {item.label}
            </button>
          )
        }
      </For>
    </div>
  )
}
