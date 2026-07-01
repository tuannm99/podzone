import { createContext, useContext } from 'solid-js'
import type { Accessor, ParentProps, Setter } from 'solid-js'

export type QueueView =
  | 'all'
  | 'my_queue'
  | 'overdue'
  | 'delivery_issues'
  | 'settlement_pending'
  | 'finance_review'

export type QueueSort = 'priority' | 'newest'
export type ShipmentSlaMode = '' | 'plus_2h' | 'plus_4h' | 'end_of_day'

export type SavedQueuePreset = {
  name: string
  queueView: QueueView
  queueSort: QueueSort
  operatorLens: string
}

export type SavedBulkTemplate = {
  name: string
  operatorAssignee: string
  shipmentSlaMode: ShipmentSlaMode
  settlementStatus: string
}

export type BulkDraft = {
  operatorAssignee: string
  shipmentSlaDueAt: string
  shipmentSlaMode: ShipmentSlaMode
  settlementStatus: string
}

export type TenantOrdersBoardContextValue = {
  activeQueueView: Accessor<QueueView>
  setActiveQueueView: Setter<QueueView>
  activeQueueSort: Accessor<QueueSort>
  setActiveQueueSort: Setter<QueueSort>
  operatorLens: Accessor<string>
  setOperatorLens: Setter<string>
  queueSearch: Accessor<string>
  setQueueSearch: Setter<string>
  applyQueueSearch: () => void
  queueViewCount: (view: QueueView) => number
  savedPresets: Accessor<SavedQueuePreset[]>
  presetName: Accessor<string>
  setPresetName: Setter<string>
  saveQueuePreset: () => void
  applyQueuePreset: (preset: SavedQueuePreset) => void
  deleteQueuePreset: (name: string) => void
  selectedOrderIDs: Accessor<string[]>
  selectVisibleOrders: () => void
  clearSelectedOrders: () => void
  bulkDraft: Accessor<BulkDraft>
  setBulkDraft: Setter<BulkDraft>
  applyRelativeShipmentSla: (mode: ShipmentSlaMode) => void
  savedBulkTemplates: Accessor<SavedBulkTemplate[]>
  bulkTemplateName: Accessor<string>
  setBulkTemplateName: Setter<string>
  saveBulkTemplate: () => void
  applyBulkTemplate: (template: SavedBulkTemplate) => void
  deleteBulkTemplate: (name: string) => void
  applyBulkUpdate: () => Promise<void>
}

const TenantOrdersBoardContext = createContext<TenantOrdersBoardContextValue>()

export function TenantOrdersBoardProvider(
  props: ParentProps<{ value: TenantOrdersBoardContextValue }>
) {
  return (
    <TenantOrdersBoardContext.Provider value={props.value}>
      {props.children}
    </TenantOrdersBoardContext.Provider>
  )
}

export function useTenantOrdersBoard() {
  const ctx = useContext(TenantOrdersBoardContext)
  if (!ctx) {
    throw new Error('TenantOrdersBoardContext is missing')
  }
  return ctx
}
