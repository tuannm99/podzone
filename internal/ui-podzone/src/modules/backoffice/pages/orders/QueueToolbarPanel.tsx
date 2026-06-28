import { For, Show } from 'solid-js'
import { Button, InputField } from '@/solid/components/common/Primitives'
import { useTenantOrdersBoard } from './board-context'

export function QueueToolbarPanel() {
  const board = useTenantOrdersBoard()

  return (
    <>
      <div class="grid gap-4 md:grid-cols-[0.7fr_1.3fr]">
        <InputField
          label="Operator lens"
          value={board.operatorLens()}
          placeholder="linh.nguyen"
          onInput={(event) => board.setOperatorLens(event.currentTarget.value)}
        />
        <div class="space-y-2">
          <p class="text-sm font-medium text-gray-700">Queue views</p>
          <div class="flex flex-wrap gap-2">
            <Button
              type="button"
              size="xs"
              color={board.activeQueueView() === 'all' ? 'blue' : 'alternative'}
              onClick={() => board.setActiveQueueView('all')}
            >
              All · {board.queueViewCount('all')}
            </Button>
            <Button
              type="button"
              size="xs"
              color={
                board.activeQueueView() === 'my_queue' ? 'blue' : 'alternative'
              }
              onClick={() => board.setActiveQueueView('my_queue')}
            >
              My queue · {board.queueViewCount('my_queue')}
            </Button>
            <Button
              type="button"
              size="xs"
              color={
                board.activeQueueView() === 'overdue' ? 'red' : 'alternative'
              }
              onClick={() => board.setActiveQueueView('overdue')}
            >
              Overdue · {board.queueViewCount('overdue')}
            </Button>
            <Button
              type="button"
              size="xs"
              color={
                board.activeQueueView() === 'delivery_issues'
                  ? 'red'
                  : 'alternative'
              }
              onClick={() => board.setActiveQueueView('delivery_issues')}
            >
              Delivery issues · {board.queueViewCount('delivery_issues')}
            </Button>
            <Button
              type="button"
              size="xs"
              color={
                board.activeQueueView() === 'settlement_pending'
                  ? 'green'
                  : 'alternative'
              }
              onClick={() => board.setActiveQueueView('settlement_pending')}
            >
              Settlement pending · {board.queueViewCount('settlement_pending')}
            </Button>
            <Button
              type="button"
              size="xs"
              color={
                board.activeQueueView() === 'finance_review'
                  ? 'red'
                  : 'alternative'
              }
              onClick={() => board.setActiveQueueView('finance_review')}
            >
              Finance review · {board.queueViewCount('finance_review')}
            </Button>
          </div>
        </div>
      </div>
      <div class="mt-4 flex flex-wrap items-center gap-2">
        <span class="text-sm font-medium text-gray-700">Sort</span>
        <Button
          type="button"
          size="xs"
          color={
            board.activeQueueSort() === 'priority' ? 'dark' : 'alternative'
          }
          onClick={() => board.setActiveQueueSort('priority')}
        >
          Priority first
        </Button>
        <Button
          type="button"
          size="xs"
          color={board.activeQueueSort() === 'newest' ? 'dark' : 'alternative'}
          onClick={() => board.setActiveQueueSort('newest')}
        >
          Newest
        </Button>
      </div>
      <div class="mt-4 rounded-lg border border-gray-200 bg-white p-4">
        <div class="grid gap-4 md:grid-cols-[0.8fr_1.2fr]">
          <InputField
            label="Save queue preset"
            value={board.presetName()}
            placeholder="Linh overdue"
            onInput={(event) => board.setPresetName(event.currentTarget.value)}
          />
          <div class="space-y-2">
            <p class="text-sm font-medium text-gray-700">Saved presets</p>
            <div class="flex flex-wrap gap-2">
              <Show
                when={board.savedPresets().length > 0}
                fallback={
                  <p class="text-sm text-gray-500">
                    No saved presets for this store yet.
                  </p>
                }
              >
                <For each={board.savedPresets()}>
                  {(preset) => (
                    <div class="flex items-center gap-2 rounded-full border border-gray-200 bg-gray-50 px-2 py-1">
                      <button
                        type="button"
                        class="text-sm font-medium text-gray-700"
                        onClick={() => board.applyQueuePreset(preset)}
                      >
                        {preset.name}
                      </button>
                      <button
                        type="button"
                        class="text-xs font-semibold text-red-600"
                        onClick={() => board.deleteQueuePreset(preset.name)}
                      >
                        remove
                      </button>
                    </div>
                  )}
                </For>
              </Show>
            </div>
          </div>
        </div>
        <div class="mt-4">
          <Button
            type="button"
            size="sm"
            color="green"
            onClick={board.saveQueuePreset}
          >
            Save current queue view
          </Button>
        </div>
      </div>
    </>
  )
}
