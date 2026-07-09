import { createEffect, onCleanup, type Accessor } from 'solid-js'

const FOCUSABLE =
    'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'

export function useFocusTrap(container: Accessor<HTMLElement | undefined>) {
    createEffect(() => {
        const el = container()
        if (!el) return

        const focusable = Array.from(el.querySelectorAll<HTMLElement>(FOCUSABLE))
        const first = focusable[0]
        const last = focusable[focusable.length - 1]

        const previouslyFocused = document.activeElement as HTMLElement | null
        first?.focus()

        const trap = (e: KeyboardEvent) => {
            if (e.key !== 'Tab') return
            if (focusable.length === 0) {
                e.preventDefault()
                return
            }
            if (e.shiftKey) {
                if (document.activeElement === first) {
                    e.preventDefault()
                    last?.focus()
                }
            } else {
                if (document.activeElement === last) {
                    e.preventDefault()
                    first?.focus()
                }
            }
        }

        const close = (e: KeyboardEvent) => {
            if (e.key === 'Escape') {
                el.dispatchEvent(new Event('focustrap:escape', { bubbles: true }))
            }
        }

        el.addEventListener('keydown', trap)
        el.addEventListener('keydown', close)

        onCleanup(() => {
            el.removeEventListener('keydown', trap)
            el.removeEventListener('keydown', close)
            previouslyFocused?.focus()
        })
    })
}
