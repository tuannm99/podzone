import { For, Show, createEffect, createMemo, createSignal } from 'solid-js'
import { Badge, Button, Card, InputField, SelectField, TextareaField } from '@/solid/components/common/Primitives'
import { Tabs } from '@/solid/components/common/Tabs'
import { IamPermissionMatrix, type IamPermissionOption, type IamPermissionSelection } from './IamPermissionMatrix'
import {
    conditionOperatorOptions,
    effectOptions,
    parseStatements,
    serializeStatements,
    type IamCondition,
    type IamStatement,
} from './iam-statement-model'

export function IamStatementBuilder(props: {
    label: string
    value: string
    onChange: (value: string) => void
    actionOptions?: IamPermissionOption[]
    builderCopy?: string
}) {
    type Mode = 'matrix' | 'builder' | 'json'
    const initialStatements = () => {
        try {
            return parseStatements(props.value)
        } catch {
            return []
        }
    }
    const [mode, setMode] = createSignal<Mode>(props.actionOptions?.length ? 'matrix' : 'builder')
    const [statements, setStatements] = createSignal<IamStatement[]>(initialStatements())
    const [parseError, setParseError] = createSignal('')

    createEffect(() => {
        try {
            const next = parseStatements(props.value)
            setStatements(next)
            setParseError('')
        } catch (error) {
            setParseError(error instanceof Error ? error.message : 'Statements JSON is invalid')
        }
    })

    const statementCount = createMemo(() => statements().length)
    const actionOptions = () => [
        { name: 'Choose a permission', value: '' },
        ...(props.actionOptions || []),
        { name: 'Custom or wildcard pattern', value: '__custom__' },
    ]
    const actionSelectValue = (value: string) => {
        if (!value) return ''
        return props.actionOptions?.some((option) => option.value === value) ? value : '__custom__'
    }
    const permissionSelection = (permission: IamPermissionOption): IamPermissionSelection => {
        const matches = statements().filter((statement) => statement.actionPattern === permission.value)
        return {
            selected: matches.length > 0,
            scoped: matches.some(
                (statement) =>
                    statement.effect !== 'allow' ||
                    statement.resourcePattern !== '*' ||
                    (statement.conditions || []).length > 0
            ),
        }
    }

    const commit = (next: IamStatement[]) => {
        setStatements(next)
        props.onChange(serializeStatements(next))
    }

    const updateStatement = (index: number, patch: Partial<IamStatement>) => {
        const next = statements().map((item, currentIndex) => (currentIndex === index ? { ...item, ...patch } : item))
        commit(next)
    }

    const removeStatement = (index: number) => {
        commit(statements().filter((_, currentIndex) => currentIndex !== index))
    }

    const addStatement = () => {
        commit([
            ...statements(),
            {
                effect: 'allow',
                actionPattern: '',
                resourcePattern: '*',
                conditions: [],
            },
        ])
    }

    const togglePermissions = (permissions: IamPermissionOption[], selected: boolean) => {
        const permissionNames = new Set(permissions.map((permission) => permission.value))
        const retained = statements().filter((statement) => !permissionNames.has(statement.actionPattern))
        if (!selected) {
            commit(retained)
            return
        }

        const existingNames = new Set(statements().map((statement) => statement.actionPattern))
        const additions = permissions
            .filter((permission) => !existingNames.has(permission.value))
            .map((permission) => ({
                effect: 'allow',
                actionPattern: permission.value,
                resourcePattern: '*',
                conditions: [],
            }))
        commit([...statements(), ...additions])
    }

    const addCondition = (index: number) => {
        const current = statements()[index]
        updateStatement(index, {
            conditions: [...(current.conditions || []), { operator: 'StringEquals', key: '', value: '' }],
        })
    }

    const updateCondition = (statementIndex: number, conditionIndex: number, patch: Partial<IamCondition>) => {
        const current = statements()[statementIndex]
        const nextConditions = (current.conditions || []).map((item, currentIndex) =>
            currentIndex === conditionIndex ? { ...item, ...patch } : item
        )
        updateStatement(statementIndex, { conditions: nextConditions })
    }

    const removeCondition = (statementIndex: number, conditionIndex: number) => {
        const current = statements()[statementIndex]
        updateStatement(statementIndex, {
            conditions: (current.conditions || []).filter((_, currentIndex) => currentIndex !== conditionIndex),
        })
    }

    const handleJsonInput = (value: string) => {
        props.onChange(value)
        try {
            const next = parseStatements(value)
            setStatements(next)
            setParseError('')
        } catch (error) {
            setParseError(error instanceof Error ? error.message : 'Statements JSON is invalid')
        }
    }

    return (
        <div class="space-y-3">
            <div class="flex flex-wrap items-center justify-between gap-3">
                <div>
                    <p class="text-sm font-medium text-gray-700">{props.label}</p>
                    <p class="mt-1 text-xs text-gray-500">
                        {props.builderCopy ||
                            'Build statements with structured fields, or switch to raw JSON for advanced edits.'}
                    </p>
                </div>
                <Badge content={`${statementCount()} statements`} color="blue" />
            </div>

            <Tabs
                value={mode()}
                items={
                    props.actionOptions?.length
                        ? [
                              { value: 'matrix', label: 'Permission matrix' },
                              { value: 'builder', label: 'Builder' },
                              { value: 'json', label: 'JSON' },
                          ]
                        : [
                              { value: 'builder', label: 'Builder' },
                              { value: 'json', label: 'JSON' },
                          ]
                }
                onChange={setMode}
            />

            <Show when={mode() === 'matrix' && props.actionOptions?.length}>
                <div class="space-y-3">
                    <p class="text-xs text-gray-500">
                        The matrix manages exact permissions. Use Builder for deny rules, resource scopes, conditions,
                        and wildcard patterns.
                    </p>
                    <IamPermissionMatrix
                        permissions={props.actionOptions || []}
                        selection={permissionSelection}
                        onToggle={(permission, selected) => togglePermissions([permission], selected)}
                        onToggleResource={(_, permissions, selected) => togglePermissions(permissions, selected)}
                    />
                </div>
            </Show>

            <Show when={mode() === 'builder'}>
                <div class="space-y-3">
                    <For each={statements()}>
                        {(statement, statementIndex) => (
                            <Card class="space-y-4 border border-gray-200 bg-gray-50 p-4 shadow-none">
                                <div class="flex flex-wrap items-center justify-between gap-3">
                                    <div class="flex flex-wrap gap-2">
                                        <Badge content={`Statement ${statementIndex() + 1}`} color="dark" />
                                        <Badge
                                            content={statement.effect}
                                            color={statement.effect === 'deny' ? 'red' : 'green'}
                                        />
                                    </div>
                                    <Button size="xs" color="red" onClick={() => removeStatement(statementIndex())}>
                                        Remove
                                    </Button>
                                </div>

                                <div class="grid gap-3 md:grid-cols-3">
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
                                    <div class="space-y-3">
                                        <SelectField
                                            label="Permission"
                                            value={actionSelectValue(statement.actionPattern)}
                                            options={actionOptions()}
                                            onChange={(event) =>
                                                updateStatement(statementIndex(), {
                                                    actionPattern:
                                                        event.currentTarget.value === '__custom__'
                                                            ? '*'
                                                            : event.currentTarget.value,
                                                })
                                            }
                                        />
                                        <Show when={actionSelectValue(statement.actionPattern) === '__custom__'}>
                                            <InputField
                                                label="Custom action pattern"
                                                value={statement.actionPattern}
                                                placeholder="orders:*"
                                                onInput={(event) =>
                                                    updateStatement(statementIndex(), {
                                                        actionPattern: event.currentTarget.value,
                                                    })
                                                }
                                            />
                                        </Show>
                                    </div>
                                    <InputField
                                        label="Resource"
                                        value={statement.resourcePattern}
                                        placeholder="*"
                                        onInput={(event) =>
                                            updateStatement(statementIndex(), {
                                                resourcePattern: event.currentTarget.value,
                                            })
                                        }
                                    />
                                </div>

                                <div class="space-y-3">
                                    <div class="flex flex-wrap items-center justify-between gap-3">
                                        <p class="text-sm font-semibold text-gray-900">Conditions</p>
                                        <Button size="xs" color="light" onClick={() => addCondition(statementIndex())}>
                                            Add condition
                                        </Button>
                                    </div>

                                    <Show
                                        when={(statement.conditions || []).length > 0}
                                        fallback={
                                            <div class="rounded-lg border border-dashed border-gray-200 bg-white px-4 py-5 text-sm text-gray-500">
                                                No conditions on this statement.
                                            </div>
                                        }
                                    >
                                        <div class="space-y-3">
                                            <For each={statement.conditions || []}>
                                                {(condition, conditionIndex) => (
                                                    <div class="grid gap-3 rounded-lg border border-gray-200 bg-white p-4 md:grid-cols-[1.2fr_1fr_1fr_auto] md:items-end">
                                                        <SelectField
                                                            label="Operator"
                                                            value={condition.operator}
                                                            options={conditionOperatorOptions}
                                                            onChange={(event) =>
                                                                updateCondition(statementIndex(), conditionIndex(), {
                                                                    operator: event.currentTarget.value,
                                                                })
                                                            }
                                                        />
                                                        <InputField
                                                            label="Key"
                                                            value={condition.key}
                                                            placeholder="principal_tag:team"
                                                            onInput={(event) =>
                                                                updateCondition(statementIndex(), conditionIndex(), {
                                                                    key: event.currentTarget.value,
                                                                })
                                                            }
                                                        />
                                                        <InputField
                                                            label="Value"
                                                            value={condition.value}
                                                            placeholder="ops"
                                                            onInput={(event) =>
                                                                updateCondition(statementIndex(), conditionIndex(), {
                                                                    value: event.currentTarget.value,
                                                                })
                                                            }
                                                        />
                                                        <Button
                                                            size="xs"
                                                            color="red"
                                                            onClick={() =>
                                                                removeCondition(statementIndex(), conditionIndex())
                                                            }
                                                        >
                                                            Remove
                                                        </Button>
                                                    </div>
                                                )}
                                            </For>
                                        </div>
                                    </Show>
                                </div>
                            </Card>
                        )}
                    </For>

                    <Button size="sm" color="dark" onClick={addStatement}>
                        Add statement
                    </Button>
                </div>
            </Show>

            <Show when={mode() === 'json'}>
                <div class="space-y-2">
                    <TextareaField
                        label="Statements JSON"
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
