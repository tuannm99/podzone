import { createEffect, createMemo, createSignal, type Accessor } from 'solid-js'

function fingerprint<T>(items: readonly T[]): string {
    if (items.length === 0) return '[]'
    return items
        .map((item, index) => {
            const row = item as Record<string, unknown>
            return String(row?.id ?? row?.key ?? index)
        })
        .join('|')
}

export function createClientPagination<T>(items: Accessor<readonly T[]>, pageSize = 8) {
    const [page, setPage] = createSignal(1)
    const total = createMemo(() => items().length)
    const totalPages = createMemo(() => Math.max(1, Math.ceil(total() / pageSize)))
    const pageItems = createMemo(() => {
        const start = (page() - 1) * pageSize
        return items().slice(start, start + pageSize)
    })

    let previousFingerprint = fingerprint(items())
    createEffect(() => {
        const fp = fingerprint(items())
        if (fp !== previousFingerprint) {
            previousFingerprint = fp
            setPage(1)
        } else if (page() > totalPages()) {
            setPage(totalPages())
        }
    })

    return {
        page,
        setPage,
        pageSize,
        total,
        pageItems,
    }
}
