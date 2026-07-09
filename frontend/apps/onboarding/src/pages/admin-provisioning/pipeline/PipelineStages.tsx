import { For } from 'solid-js'
import type { StoreRequest } from '@podzone/shared/services/onboarding'
import type { StoreRequestTransition } from '@podzone/shared/services/onboarding/provisioning'
import { classes } from '@podzone/shared/ui/shared/utils'

type PipelineStage = {
    key: string
    label: string
    steps: string[]
}

const stages: PipelineStage[] = [
    { key: 'request', label: 'Request', steps: ['request.created'] },
    {
        key: 'approval',
        label: 'Approval',
        steps: ['approval.queued', 'request.retried'],
    },
    {
        key: 'planning',
        label: 'Plan',
        steps: ['planning.started', 'planning.completed'],
    },
    {
        key: 'provisioning',
        label: 'Provision',
        steps: ['provisioning.started', 'provisioning.failed'],
    },
    { key: 'route', label: 'Publish route', steps: ['route.ready'] },
    { key: 'finalize', label: 'Finalize store', steps: ['store.finalized'] },
    { key: 'ready', label: 'Ready', steps: ['request.ready'] },
]

function fallbackStageIndex(status: StoreRequest['status']) {
    if (status === 'ready') return stages.length - 1
    if (status === 'provisioning') return 3
    if (status === 'planning' || status === 'planned') return 2
    if (status === 'queued' || status === 'pending_approval') return 1
    return 0
}

export function PipelineStages(props: { request: StoreRequest; transitions: StoreRequestTransition[] }) {
    const completedSteps = () => new Set(props.transitions.map((transition) => transition.step))
    const currentIndex = () => {
        if (props.request.status === 'ready') return stages.length
        let index = fallbackStageIndex(props.request.status)
        for (const [stageIndex, stage] of stages.entries()) {
            if (stage.steps.some((step) => completedSteps().has(step))) {
                index = Math.max(index, stageIndex + 1)
            }
        }
        return Math.min(index, stages.length - 1)
    }
    const isFailed = () => props.request.status.startsWith('failed')

    return (
        <div class="overflow-x-auto pb-2">
            <ol class="grid min-w-[760px] grid-cols-7">
                <For each={stages}>
                    {(stage, index) => {
                        const completed = () => props.request.status === 'ready' || index() < currentIndex()
                        const active = () => props.request.status !== 'ready' && index() === currentIndex()
                        return (
                            <li class="relative pr-3 last:pr-0">
                                <div
                                    class={classes(
                                        'absolute left-7 right-0 top-3 h-px',
                                        completed() ? 'bg-emerald-500' : 'bg-gray-200',
                                        index() === stages.length - 1 && 'hidden'
                                    )}
                                />
                                <div class="relative flex flex-col gap-2">
                                    <span
                                        class={classes(
                                            'flex size-7 items-center justify-center rounded-full border text-xs font-semibold',
                                            completed()
                                                ? 'border-emerald-600 bg-emerald-600 text-white'
                                                : active() && isFailed()
                                                  ? 'border-red-600 bg-red-50 text-red-700'
                                                  : active()
                                                    ? 'border-gray-950 bg-gray-950 text-white'
                                                    : 'border-gray-300 bg-white text-gray-500'
                                        )}
                                    >
                                        {index() + 1}
                                    </span>
                                    <span
                                        class={classes(
                                            'max-w-24 text-xs font-medium',
                                            active() ? 'text-gray-950' : 'text-gray-500'
                                        )}
                                    >
                                        {stage.label}
                                    </span>
                                </div>
                            </li>
                        )
                    }}
                </For>
            </ol>
        </div>
    )
}
