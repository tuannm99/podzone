import { For, Show, createEffect, createSignal, onCleanup, type JSX } from 'solid-js'
import { classes } from '../../shared/utils'

export type MegaMenuLink = {
  label: string
  href: string
  description?: string
  icon?: JSX.Element
}

export type MegaMenuSection = {
  title: string
  links: MegaMenuLink[]
}

export type MegaMenuItem = {
  label: string
  href?: string
  active?: boolean
  sections?: MegaMenuSection[]
}

export function MegaMenu(props: {
  brand?: JSX.Element
  items: MegaMenuItem[]
  actions?: JSX.Element
  class?: string
}) {
  const [openIndex, setOpenIndex] = createSignal<number | null>(null)
  let container: HTMLDivElement | undefined

  createEffect(() => {
    const current = openIndex()
    if (current === null) return

    const handlePointerDown = (event: MouseEvent) => {
      if (container && !container.contains(event.target as Node)) {
        setOpenIndex(null)
      }
    }

    document.addEventListener('mousedown', handlePointerDown)
    onCleanup(() => {
      document.removeEventListener('mousedown', handlePointerDown)
    })
  })

  return (
    <div
      class={classes(
        'relative rounded-3xl border border-gray-200 bg-white px-5 py-4 shadow-sm',
        props.class
      )}
      ref={container}
    >
      <div class="flex flex-col gap-4 xl:flex-row xl:items-center xl:justify-between">
        <div class="flex items-center justify-between gap-4">
          <Show when={props.brand}>
            <div>{props.brand}</div>
          </Show>
          <Show when={props.actions}>
            <div class="xl:hidden">{props.actions}</div>
          </Show>
        </div>

        <nav class="flex flex-wrap items-center gap-2">
          <For each={props.items}>
            {(item, index) =>
              item.sections?.length ? (
                <button
                  type="button"
                  class={classes(
                    'inline-flex items-center gap-2 rounded-xl px-3 py-2 text-sm font-medium transition',
                    item.active || openIndex() === index()
                      ? 'bg-blue-50 text-blue-900'
                      : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'
                  )}
                  onClick={() => setOpenIndex((value) => (value === index() ? null : index()))}
                >
                  <span>{item.label}</span>
                  <span
                    class={classes(
                      'text-gray-400 transition',
                      openIndex() === index() && 'rotate-180'
                    )}
                  >
                    ⌄
                  </span>
                </button>
              ) : (
                <a
                  href={item.href}
                  class={classes(
                    'inline-flex items-center rounded-xl px-3 py-2 text-sm font-medium transition',
                    item.active
                      ? 'bg-blue-700 text-white'
                      : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'
                  )}
                >
                  {item.label}
                </a>
              )
            }
          </For>
        </nav>

        <Show when={props.actions}>
          <div class="hidden xl:block">{props.actions}</div>
        </Show>
      </div>

      <Show when={openIndex() !== null}>
        <div class="absolute inset-x-4 top-full z-40 mt-3 rounded-3xl border border-gray-200 bg-white p-6 shadow-2xl">
          <For each={props.items[openIndex() ?? 0]?.sections ?? []}>
            {(section) => (
              <div class="mb-6 last:mb-0">
                <p class="mb-3 text-xs font-semibold uppercase tracking-[0.22em] text-gray-400">
                  {section.title}
                </p>
                <div class="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
                  <For each={section.links}>
                    {(link) => (
                      <a
                        href={link.href}
                        class="rounded-2xl border border-gray-200 px-4 py-3 transition hover:border-blue-200 hover:bg-blue-50"
                        onClick={() => setOpenIndex(null)}
                      >
                        <div class="flex items-start gap-3">
                          <Show when={link.icon}>
                            <div class="pt-0.5 text-blue-700">{link.icon}</div>
                          </Show>
                          <div class="space-y-1">
                            <p class="text-sm font-semibold text-gray-900">{link.label}</p>
                            <Show when={link.description}>
                              <p class="text-sm leading-6 text-gray-500">{link.description}</p>
                            </Show>
                          </div>
                        </div>
                      </a>
                    )}
                  </For>
                </div>
              </div>
            )}
          </For>
        </div>
      </Show>
    </div>
  )
}
