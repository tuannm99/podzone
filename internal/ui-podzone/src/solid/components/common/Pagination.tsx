import { For, Show, createMemo } from 'solid-js'
import { classes } from '../../shared/utils'
import { Button } from './Primitives'

type PaginationProps = {
  page: number
  pageSize: number
  total: number
  loading?: boolean
  class?: string
  onPageChange: (page: number) => void
}

function buildPageItems(page: number, totalPages: number) {
  if (totalPages <= 7) {
    return Array.from({ length: totalPages }, (_, index) => index + 1)
  }

  const pages = new Set<number>([1, totalPages, page - 1, page, page + 1])
  const visible = [...pages]
    .filter((value) => value >= 1 && value <= totalPages)
    .sort((a, b) => a - b)
  const items: Array<number | 'ellipsis'> = []

  for (const value of visible) {
    const previous = items[items.length - 1]
    if (typeof previous === 'number' && value - previous > 1) {
      items.push('ellipsis')
    }
    items.push(value)
  }

  return items
}

export function Pagination(props: PaginationProps) {
  const totalPages = createMemo(() =>
    Math.max(1, Math.ceil(props.total / props.pageSize))
  )
  const start = createMemo(() =>
    props.total === 0 ? 0 : (props.page - 1) * props.pageSize + 1
  )
  const finish = createMemo(() =>
    Math.min(props.page * props.pageSize, props.total)
  )
  const items = createMemo(() => buildPageItems(props.page, totalPages()))
  const changePage = (event: MouseEvent, page: number) => {
    event.preventDefault()
    event.stopPropagation()
    if (props.loading || page === props.page) return
    props.onPageChange(page)
  }

  return (
    <Show when={totalPages() > 1}>
      <div
        class={classes(
          'flex flex-col gap-4 border-t border-gray-100 pt-4 md:flex-row md:items-center md:justify-between',
          props.class
        )}
      >
        <p class="text-sm text-gray-500">
          Showing {start()}-{finish()} of {props.total}
        </p>

        <div class="flex flex-wrap items-center gap-2">
          <Button
            type="button"
            pill
            size="xs"
            color="alternative"
            disabled={props.loading || props.page <= 1}
            onClick={(event) => changePage(event, Math.max(1, props.page - 1))}
          >
            Previous
          </Button>

          <div class="flex flex-wrap items-center gap-2">
            <For each={items()}>
              {(item) =>
                item === 'ellipsis' ? (
                  <span class="px-2 text-sm text-gray-400">...</span>
                ) : (
                  <Button
                    type="button"
                    pill
                    size="xs"
                    color={item === props.page ? 'dark' : 'alternative'}
                    loading={props.loading && item === props.page}
                    disabled={props.loading}
                    aria-current={item === props.page ? 'page' : undefined}
                    onClick={(event) => changePage(event, item)}
                  >
                    {item}
                  </Button>
                )
              }
            </For>
          </div>

          <Button
            type="button"
            pill
            size="xs"
            color="alternative"
            disabled={props.loading || props.page >= totalPages()}
            onClick={(event) =>
              changePage(event, Math.min(totalPages(), props.page + 1))
            }
          >
            Next
          </Button>
        </div>
      </div>
    </Show>
  )
}
