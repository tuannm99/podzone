import { For } from 'solid-js'
import { toasts, toaster, type ToastType } from './createToaster'
import { classes } from '../shared/utils'

const toastColorClasses: Record<ToastType, string> = {
    success: 'bg-green-700 text-white',
    error: 'bg-red-700 text-white',
    info: 'bg-gray-800 text-white',
    warning: 'bg-yellow-600 text-white',
}

export function ToasterProvider() {
    return (
        <div aria-live="polite" aria-atomic="false" class="fixed bottom-4 right-4 z-50 flex flex-col gap-2">
            <For each={toasts()}>
                {(toast) => (
                    <div
                        role="status"
                        class={classes(
                            'flex items-center gap-3 rounded-lg px-4 py-3 shadow-lg text-sm font-medium',
                            toastColorClasses[toast.type]
                        )}
                    >
                        <span class="flex-1">{toast.message}</span>
                        <button
                            type="button"
                            aria-label="Dismiss"
                            class="ml-2 opacity-70 hover:opacity-100"
                            onClick={() => toaster.dismiss(toast.id)}
                        >
                            ✕
                        </button>
                    </div>
                )}
            </For>
        </div>
    )
}
