import { createEffect, createMemo, createSignal, type Accessor } from 'solid-js'

export function createClientPagination<T>(
  items: Accessor<readonly T[]>,
  pageSize = 8
) {
  const [page, setPage] = createSignal(1)
  const total = createMemo(() => items().length)
  const totalPages = createMemo(() =>
    Math.max(1, Math.ceil(total() / pageSize))
  )
  const pageItems = createMemo(() => {
    const start = (page() - 1) * pageSize
    return items().slice(start, start + pageSize)
  })
  let previousItems = items()

  createEffect(() => {
    const currentItems = items()
    if (currentItems !== previousItems) {
      previousItems = currentItems
      setPage(1)
      return
    }
    if (page() > totalPages()) {
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
