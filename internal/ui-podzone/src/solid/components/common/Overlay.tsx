import {
  For,
  Show,
  createEffect,
  createSignal,
  onCleanup,
  type JSX,
  type ParentProps
} from 'solid-js'
import { classes } from '../../shared/utils'

type OverlayPosition = 'top' | 'right' | 'bottom' | 'left'

const floatingPositionClasses: Record<OverlayPosition, string> = {
  top: 'bottom-full left-1/2 mb-2 -translate-x-1/2',
  right: 'left-full top-1/2 ml-2 -translate-y-1/2',
  bottom: 'left-1/2 top-full mt-2 -translate-x-1/2',
  left: 'right-full top-1/2 mr-2 -translate-y-1/2'
}

export type DropdownItem = {
  label: string
  href?: string
  onSelect?: () => void
  tone?: 'default' | 'danger'
}

export function DropdownMenu(props: {
  label?: string
  trigger?: JSX.Element
  items: DropdownItem[]
  class?: string
  menuClass?: string
}) {
  const [open, setOpen] = createSignal(false)
  let container: HTMLDivElement | undefined

  createEffect(() => {
    if (!open()) return

    const handlePointerDown = (event: MouseEvent) => {
      if (container && !container.contains(event.target as Node)) {
        setOpen(false)
      }
    }

    document.addEventListener('mousedown', handlePointerDown)
    onCleanup(() => {
      document.removeEventListener('mousedown', handlePointerDown)
    })
  })

  return (
    <div class={classes('relative inline-block text-left', props.class)} ref={container}>
      <button
        type="button"
        class="inline-flex items-center gap-2 rounded-xl border border-gray-200 bg-white px-4 py-2.5 text-sm font-medium text-gray-700 shadow-sm transition hover:bg-gray-50"
        onClick={() => setOpen((value) => !value)}
      >
        {props.trigger ?? props.label ?? 'More'}
        <span class={classes('text-gray-400 transition', open() && 'rotate-180')}>⌄</span>
      </button>

      <Show when={open()}>
        <div
          class={classes(
            'absolute right-0 z-30 mt-2 min-w-48 rounded-2xl border border-gray-200 bg-white p-2 shadow-xl',
            props.menuClass
          )}
        >
          <div class="space-y-1">
            <For each={props.items}>
              {(item) =>
                item.href ? (
                  <a
                    href={item.href}
                    class={classes(
                      'block rounded-xl px-3 py-2 text-sm transition hover:bg-gray-50',
                      item.tone === 'danger' ? 'text-red-600' : 'text-gray-700'
                    )}
                    onClick={() => setOpen(false)}
                  >
                    {item.label}
                  </a>
                ) : (
                  <button
                    type="button"
                    class={classes(
                      'block w-full rounded-xl px-3 py-2 text-left text-sm transition hover:bg-gray-50',
                      item.tone === 'danger' ? 'text-red-600' : 'text-gray-700'
                    )}
                    onClick={() => {
                      item.onSelect?.()
                      setOpen(false)
                    }}
                  >
                    {item.label}
                  </button>
                )
              }
            </For>
          </div>
        </div>
      </Show>
    </div>
  )
}

export function Tooltip(
  props: ParentProps<{
    content: JSX.Element
    position?: OverlayPosition
    class?: string
    panelClass?: string
  }>
) {
  return (
    <div class={classes('group relative inline-flex', props.class)}>
      {props.children}
      <div
        class={classes(
          'pointer-events-none absolute z-30 rounded-lg bg-gray-900 px-2.5 py-1.5 text-xs text-white opacity-0 shadow-lg transition group-hover:opacity-100',
          floatingPositionClasses[props.position ?? 'top'],
          props.panelClass
        )}
      >
        {props.content}
      </div>
    </div>
  )
}

export function Popover(
  props: ParentProps<{
    content: JSX.Element
    position?: OverlayPosition
    class?: string
    panelClass?: string
  }>
) {
  return (
    <div class={classes('group relative inline-flex', props.class)}>
      {props.children}
      <div
        class={classes(
          'pointer-events-none absolute z-30 min-w-56 rounded-2xl border border-gray-200 bg-white p-4 text-sm text-gray-600 opacity-0 shadow-xl transition group-hover:pointer-events-auto group-hover:opacity-100',
          floatingPositionClasses[props.position ?? 'bottom'],
          props.panelClass
        )}
      >
        {props.content}
      </div>
    </div>
  )
}

type ModalSize = 'sm' | 'md' | 'lg' | 'xl'

const modalSizeClasses: Record<ModalSize, string> = {
  sm: 'max-w-md',
  md: 'max-w-2xl',
  lg: 'max-w-4xl',
  xl: 'max-w-6xl'
}

export function Modal(
  props: ParentProps<{
    open: boolean
    title?: string
    footer?: JSX.Element
    size?: ModalSize
    class?: string
    onClose?: () => void
  }>
) {
  return (
    <Show when={props.open}>
      <div
        class="fixed inset-0 z-50 flex items-center justify-center bg-gray-900/40 p-4 backdrop-blur-sm"
        onClick={(event) => {
          if (event.target === event.currentTarget) {
            props.onClose?.()
          }
        }}
      >
        <div
          class={classes(
            'w-full rounded-3xl border border-gray-200 bg-white p-6 shadow-2xl',
            modalSizeClasses[props.size ?? 'md'],
            props.class
          )}
        >
          <div class="flex items-start justify-between gap-4">
            <Show when={props.title}>
              <h2 class="text-xl font-semibold tracking-tight text-gray-900">{props.title}</h2>
            </Show>
            <Show when={props.onClose}>
              <button
                type="button"
                class="rounded-full px-2 py-1 text-sm font-medium text-gray-500 transition hover:bg-gray-100 hover:text-gray-900"
                onClick={() => props.onClose?.()}
                aria-label="Close modal"
              >
                ✕
              </button>
            </Show>
          </div>
          <div class="mt-4">{props.children}</div>
          <Show when={props.footer}>
            <div class="mt-6 flex flex-wrap justify-end gap-3">{props.footer}</div>
          </Show>
        </div>
      </div>
    </Show>
  )
}

type DrawerSide = 'left' | 'right'

const drawerSideClasses: Record<DrawerSide, string> = {
  left: 'left-0',
  right: 'right-0'
}

export function Drawer(
  props: ParentProps<{
    open: boolean
    title?: string
    side?: DrawerSide
    class?: string
    onClose?: () => void
  }>
) {
  return (
    <Show when={props.open}>
      <div
        class="fixed inset-0 z-50 bg-gray-900/30 backdrop-blur-sm"
        onClick={(event) => {
          if (event.target === event.currentTarget) {
            props.onClose?.()
          }
        }}
      >
        <aside
          class={classes(
            'absolute top-0 h-full w-full max-w-md overflow-y-auto border-gray-200 bg-white p-6 shadow-2xl',
            drawerSideClasses[props.side ?? 'right'],
            props.side === 'left' ? 'border-r' : 'border-l',
            props.class
          )}
        >
          <div class="flex items-start justify-between gap-4">
            <Show when={props.title}>
              <h2 class="text-lg font-semibold text-gray-900">{props.title}</h2>
            </Show>
            <Show when={props.onClose}>
              <button
                type="button"
                class="rounded-full px-2 py-1 text-sm font-medium text-gray-500 transition hover:bg-gray-100 hover:text-gray-900"
                onClick={() => props.onClose?.()}
                aria-label="Close drawer"
              >
                ✕
              </button>
            </Show>
          </div>
          <div class="mt-5">{props.children}</div>
        </aside>
      </div>
    </Show>
  )
}
