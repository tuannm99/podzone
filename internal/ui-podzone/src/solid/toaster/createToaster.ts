import { createSignal } from 'solid-js'

export type ToastType = 'success' | 'error' | 'info' | 'warning'

export type Toast = {
    id: string
    type: ToastType
    message: string
}

let counter = 0
const [toasts, setToasts] = createSignal<Toast[]>([])

function push(type: ToastType, message: string, durationMs = 4000) {
    const id = `toast-${++counter}`
    setToasts((prev) => [...prev, { id, type, message }])
    setTimeout(() => dismiss(id), durationMs)
}

function dismiss(id: string) {
    setToasts((prev) => prev.filter((t) => t.id !== id))
}

export const toaster = {
    success: (msg: string, ms?: number) => push('success', msg, ms),
    error: (msg: string, ms?: number) => push('error', msg, ms),
    info: (msg: string, ms?: number) => push('info', msg, ms),
    warning: (msg: string, ms?: number) => push('warning', msg, ms),
    dismiss,
}

export function useToaster() {
    return toaster
}

export { toasts }
