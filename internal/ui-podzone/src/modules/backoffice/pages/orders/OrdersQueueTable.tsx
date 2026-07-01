import { For, type Accessor, type Setter } from 'solid-js'
import type { PageInfo } from '@/services/collection'
import type { RoutedOrder } from '@/services/orders'
import {
  DataTable,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from '@/solid/components/common/DataTable'
import { Pagination } from '@/solid/components/common/Pagination'
import { Badge, Button } from '@/solid/components/common/Primitives'

type BadgeColor =
  | 'blue'
  | 'indigo'
  | 'green'
  | 'yellow'
  | 'pink'
  | 'dark'
  | 'red'

type OrdersQueueTableProps = {
  orders: Accessor<RoutedOrder[]>
  pageInfo: Accessor<PageInfo>
  page: Accessor<number>
  onPageChange: (page: number) => void
  detailOrderID: Accessor<string>
  setDetailOrderID: Setter<string>
  isSelected: (orderID: string) => boolean
  toggleSelected: (orderID: string, selected: boolean) => void
  priorityScoreFor: (order: RoutedOrder) => number
  statusColor: (status: string) => BadgeColor
  exceptionColor: (status: string) => BadgeColor
  settlementColor: (status: string) => BadgeColor
}

export function OrdersQueueTable(props: OrdersQueueTableProps) {
  return (
    <>
      <DataTable>
        <TableHead>
          <TableRow>
            <TableHeaderCell class="w-10">
              <span class="sr-only">Select</span>
            </TableHeaderCell>
            <TableHeaderCell>Order</TableHeaderCell>
            <TableHeaderCell>Partner</TableHeaderCell>
            <TableHeaderCell>Priority</TableHeaderCell>
            <TableHeaderCell>Execution</TableHeaderCell>
            <TableHeaderCell>Settlement</TableHeaderCell>
            <TableHeaderCell class="text-right">Action</TableHeaderCell>
          </TableRow>
        </TableHead>
        <TableBody>
          <For each={props.orders()}>
            {(order) => (
              <TableRow>
                <TableCell>
                  <input
                    type="checkbox"
                    class="size-4 rounded border-gray-300 text-gray-950"
                    checked={props.isSelected(order.id)}
                    aria-label={`Select order ${order.id}`}
                    onChange={(event) =>
                      props.toggleSelected(
                        order.id,
                        event.currentTarget.checked
                      )
                    }
                  />
                </TableCell>
                <TableCell>
                  <p class="font-semibold text-gray-900">
                    {order.productTitle}
                  </p>
                  <p class="mt-1 max-w-44 truncate text-xs text-gray-500">
                    {order.id}
                  </p>
                  <p class="mt-1 text-xs text-gray-500">
                    {order.customerName} · {order.quantity} item
                    {order.quantity === 1 ? '' : 's'}
                  </p>
                </TableCell>
                <TableCell class="text-gray-700">
                  {order.partner || 'Unassigned'}
                </TableCell>
                <TableCell>
                  <Badge
                    content={String(props.priorityScoreFor(order))}
                    color="indigo"
                  />
                </TableCell>
                <TableCell>
                  <div class="flex max-w-44 flex-wrap gap-1">
                    <Badge
                      content={order.status.replaceAll('_', ' ')}
                      color={props.statusColor(order.status)}
                    />
                    {order.exceptionStatus ? (
                      <Badge
                        content={order.exceptionStatus.replaceAll('_', ' ')}
                        color={props.exceptionColor(order.exceptionStatus)}
                      />
                    ) : null}
                  </div>
                </TableCell>
                <TableCell>
                  <Badge
                    content={
                      order.settlementStatus?.replaceAll('_', ' ') || 'pending'
                    }
                    color={props.settlementColor(order.settlementStatus)}
                  />
                </TableCell>
                <TableCell class="text-right">
                  <Button
                    size="xs"
                    color={
                      props.detailOrderID() === order.id
                        ? 'blue'
                        : 'alternative'
                    }
                    onClick={() => props.setDetailOrderID(order.id)}
                  >
                    {props.detailOrderID() === order.id ? 'Open' : 'Inspect'}
                  </Button>
                </TableCell>
              </TableRow>
            )}
          </For>
        </TableBody>
      </DataTable>
      <Pagination
        page={props.page()}
        pageSize={props.pageInfo().pageSize}
        total={props.pageInfo().total}
        onPageChange={props.onPageChange}
      />
    </>
  )
}
