import { For, Show, createEffect, createMemo, createSignal } from 'solid-js'
import type { RoleTrustStatement } from '../../../services/iam'
import { Tabs } from './Tabs'
import { Badge, Button, Card, InputField, SelectField, TextareaField } from './Primitives'

const effectOptions = [
    { name: 'Allow', value: 'allow' },
    { name: 'Deny', value: 'deny' },
]

const principalTypeOptions = [
    { name: 'User', value: 'user' },
    { name: 'Platform role', value: 'platform_role' },
    { name: 'Tenant role', value: 'tenant_role' },
    { name: 'Service', value: 'service' },
]

function normalizeTrustStatements(raw: string): RoleTrustStatement[] {
    try {
        const parsed = JSON.parse(raw || '[]')
        if (!Array.isArray(parsed)) return []
        return parsed.map((item) => ({
            effect: typeof item?.effect === 'string' ? item.effect : 'allow',
            principalType: typeof item?.principalType === 'string' ? item.principalType : 'user',
            principalPattern: typeof item?.principalPattern === 'string' ? item.principalPattern : '',
            tenantPattern: typeof item?.tenantPattern === 'string' ? item.tenantPattern : '*',
            externalIdPattern: typeof item?.externalIdPattern === 'string' ? item.externalIdPattern : '',
        }))
    } catch {
        return []
    }
}

function serializeTrustStatements(items: RoleTrustStatement[]) {
    return JSON.stringify(items, null, 2)
}

export function IamTrustPolicyBuilder(props: { label: string; value: string; onChange: (value: string) => void }) {
    const [mode, setMode] = createSignal<'builder' | 'json'>('builder')
    const [statements, setStatements] = createSignal<RoleTrustStatement[]>(normalizeTrustStatements(props.value))
    const [parseError, setParseError] = createSignal('')

    createEffect(() => {
        const next = normalizeTrustStatements(props.value)
        setStatements(next)
        setParseError('')
    })

    const count = createMemo(() => statements().length)

    const commit = (next: RoleTrustStatement[]) => {
        setStatements(next)
        props.onChange(serializeTrustStatements(next))
    }

    const updateStatement = (index: number, patch: Partial<RoleTrustStatement>) => {
        commit(statements().map((item, currentIndex) => (currentIndex === index ? { ...item, ...patch } : item)))
    }

    const removeStatement = (index: number) => {
        commit(statements().filter((_, currentIndex) => currentIndex !== index))
    }

    const addStatement = () => {
        commit([
            ...statements(),
            {
                effect: 'allow',
                principalType: 'user',
                principalPattern: '',
                tenantPattern: '*',
                externalIdPattern: '',
            },
        ])
    }

    const handleJsonInput = (value: string) => {
        props.onChange(value)
        try {
            setStatements(normalizeTrustStatements(value))
            setParseError('')
        } catch {
            setParseError('Invalid JSON')
        }
    }

    return (
        <div class="space-y-3">
            <div class="flex flex-wrap items-center justify-between gap-3">
                <div>
                    <p class="text-sm font-medium text-gray-700">{props.label}</p>
                    <p class="mt-1 text-xs text-gray-500">
                        Define who may assume the role, with tenant and external ID patterns.
                    </p>
                </div>
                <Badge content={`${count()} trust statements`} color="blue" />
            </div>

            <Tabs
                value={mode()}
                items={[
                    { value: 'builder', label: 'Builder' },
                    { value: 'json', label: 'JSON' },
                ]}
                onChange={(value) => setMode(value as 'builder' | 'json')}
            />

            <Show when={mode() === 'builder'}>
                <div class="space-y-3">
                    <For each={statements()}>
                        {(statement, statementIndex) => (
                            <Card class="space-y-4 border border-gray-200 bg-gray-50 p-4 shadow-none">
                                <div class="flex flex-wrap items-center justify-between gap-3">
                                    <div class="flex flex-wrap gap-2">
                                        <Badge content={`Trust ${statementIndex() + 1}`} color="dark" />
                                        <Badge
                                            content={statement.effect}
                                            color={statement.effect === 'deny' ? 'red' : 'green'}
                                        />
                                        <Badge content={statement.principalType} color="blue" />
                                    </div>
                                    <Button size="xs" color="red" onClick={() => removeStatement(statementIndex())}>
                                        Remove
                                    </Button>
                                </div>

                                <div class="grid gap-3 md:grid-cols-2">
                                    <SelectField
                                        label="Effect"
                                        value={statement.effect}
                                        options={effectOptions}
                                        onChange={(event) =>
                                            updateStatement(statementIndex(), {
                                                effect: event.currentTarget.value,
                                            })
                                        }
                                    />
                                    <SelectField
                                        label="Principal type"
                                        value={statement.principalType}
                                        options={principalTypeOptions}
                                        onChange={(event) =>
                                            updateStatement(statementIndex(), {
                                                principalType: event.currentTarget.value,
                                            })
                                        }
                                    />
                                </div>

                                <div class="grid gap-3 md:grid-cols-3">
                                    <InputField
                                        label="Principal pattern"
                                        value={statement.principalPattern}
                                        placeholder="backoffice.podzone.internal"
                                        onInput={(event) =>
                                            updateStatement(statementIndex(), {
                                                principalPattern: event.currentTarget.value,
                                            })
                                        }
                                    />
                                    <InputField
                                        label="Tenant pattern"
                                        value={statement.tenantPattern || ''}
                                        placeholder="*"
                                        onInput={(event) =>
                                            updateStatement(statementIndex(), {
                                                tenantPattern: event.currentTarget.value,
                                            })
                                        }
                                    />
                                    <InputField
                                        label="External ID pattern"
                                        value={statement.externalIdPattern || ''}
                                        placeholder="partner-*"
                                        onInput={(event) =>
                                            updateStatement(statementIndex(), {
                                                externalIdPattern: event.currentTarget.value,
                                            })
                                        }
                                    />
                                </div>
                            </Card>
                        )}
                    </For>

                    <Button size="sm" color="dark" onClick={addStatement}>
                        Add trust statement
                    </Button>
                </div>
            </Show>

            <Show when={mode() === 'json'}>
                <div class="space-y-2">
                    <TextareaField
                        label="Trust policy JSON"
                        rows={8}
                        value={props.value}
                        onInput={(event) => handleJsonInput(event.currentTarget.value)}
                    />
                    <Show when={parseError()}>
                        <p class="text-sm text-red-600">{parseError()}</p>
                    </Show>
                </div>
            </Show>
        </div>
    )
}
