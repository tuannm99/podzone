import { For, Show } from 'solid-js'
import { EmptyBlock } from '@podzone/shared/ui/components/common/Feedback'
import { Badge } from '@podzone/shared/ui/components/common/Primitives'
import { classes } from '@podzone/shared/ui/shared/utils'
import type { SimulateMatchedStatement } from '@podzone/shared/services/iam'
import { useAdminIamTrustSim } from './context'

function formatConditions(conditions: Array<{ operator: string; key: string; value: string }>) {
    return conditions.map((condition) => `${condition.operator} ${condition.key}=${condition.value}`).join(' · ')
}

function StatementMatchCard(props: { statement: SimulateMatchedStatement; showPolicyName?: boolean }) {
    const trust = useAdminIamTrustSim()

    return (
        <div class="rounded-md bg-gray-50 p-3 text-xs text-gray-600">
            <div class="flex flex-wrap items-center gap-2">
                <Badge
                    content={props.statement.effect}
                    color={props.statement.effect.toLowerCase() === 'deny' ? 'red' : 'green'}
                />
                <Badge
                    content={trust.statementSourceLabel(props.statement.source)}
                    color={trust.simulationSourceColor(props.statement.source)}
                />
                <Show when={props.showPolicyName ? true : props.statement.policyName}>
                    <Badge content={props.statement.policyName || 'inline'} color="dark" />
                </Show>
            </div>
            <p class="mt-2">
                {props.statement.actionPattern} on {props.statement.resourcePattern}
            </p>
            <Show when={(props.statement.conditions || []).length > 0}>
                <p class="mt-2 text-[11px] text-gray-500">
                    Conditions: {formatConditions(props.statement.conditions || [])}
                </p>
            </Show>
        </div>
    )
}

export function SimulationResultSection() {
    const trust = useAdminIamTrustSim()

    return (
        <Show
            when={trust.simulation()}
            fallback={
                <EmptyBlock
                    title="No simulation yet"
                    copy="Run a simulation to inspect why a request is allowed or denied across identity, boundaries, session policy, and SCP layers."
                />
            }
        >
            {(result) => (
                <div class="space-y-4 rounded-lg border border-gray-200 bg-gray-50 p-4">
                    <div class="flex flex-wrap items-center gap-3">
                        <Badge
                            content={result().allowed ? 'allowed' : 'denied'}
                            color={result().allowed ? 'green' : 'red'}
                        />
                        <Badge
                            content={result().decisionSource}
                            color={trust.simulationSourceColor(result().decisionSource)}
                        />
                        <Badge content={`${result().layers?.length || 0} layers`} color="dark" />
                        <Badge content={`${result().matchedStatements?.length || 0} top matches`} color="blue" />
                    </div>
                    <p class="text-sm text-gray-600">{result().reason}</p>
                    <Show when={(result().matchedStatements || []).length > 0}>
                        <div class="rounded-lg border border-gray-200 bg-white p-4">
                            <div class="flex flex-wrap items-center gap-2">
                                <Badge content="decision matches" color="dark" />
                                <Badge
                                    content={
                                        result().matchedStatements?.some(
                                            (statement) => statement.effect.toLowerCase() === 'deny'
                                        )
                                            ? 'explicit deny present'
                                            : 'allow path'
                                    }
                                    color={
                                        result().matchedStatements?.some(
                                            (statement) => statement.effect.toLowerCase() === 'deny'
                                        )
                                            ? 'red'
                                            : 'green'
                                    }
                                />
                            </div>
                            <div class="mt-3 space-y-2">
                                <For each={result().matchedStatements || []}>
                                    {(statement) => <StatementMatchCard statement={statement} />}
                                </For>
                            </div>
                        </div>
                    </Show>
                    <div class="space-y-3">
                        <For each={result().layers || []}>
                            {(layer) => (
                                <div
                                    class={classes(
                                        'rounded-lg border p-4',
                                        trust.simulationLayerTone(layer.allowed, layer.reason)
                                    )}
                                >
                                    <div class="flex flex-wrap items-center gap-2">
                                        <Badge content={layer.layer} color={trust.simulationSourceColor(layer.layer)} />
                                        <Badge
                                            content={layer.allowed ? 'allowed' : 'denied'}
                                            color={layer.allowed ? 'green' : 'red'}
                                        />
                                        <Show when={layer.reason.toLowerCase().includes('deny')}>
                                            <Badge content="explicit deny" color="red" />
                                        </Show>
                                        <Show when={layer.reason.toLowerCase().includes('boundary')}>
                                            <Badge content="boundary gate" color="pink" />
                                        </Show>
                                        <Show when={layer.reason.toLowerCase().includes('scp')}>
                                            <Badge content="scp gate" color="yellow" />
                                        </Show>
                                        <Show when={layer.reason.toLowerCase().includes('session policy')}>
                                            <Badge content="session scope-down" color="indigo" />
                                        </Show>
                                    </div>
                                    <p class="mt-2 text-sm text-gray-600">{layer.reason}</p>
                                    <Show when={(layer.matchedStatements || []).length > 0}>
                                        <div class="mt-3 space-y-2">
                                            <For each={layer.matchedStatements || []}>
                                                {(statement) => (
                                                    <StatementMatchCard statement={statement} showPolicyName />
                                                )}
                                            </For>
                                        </div>
                                    </Show>
                                </div>
                            )}
                        </For>
                    </div>
                </div>
            )}
        </Show>
    )
}
