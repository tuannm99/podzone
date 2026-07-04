import { For, Show, createEffect, createMemo, createSignal } from 'solid-js'
import { Tabs } from './Tabs'
import {
  Badge,
  Button,
  Card,
  InputField,
  SelectField,
  TextareaField,
  type SelectOption,
} from './Primitives'

type Condition = {
  operator: string
  key: string
  value: string
}

type Statement = {
  effect: string
  actionPattern: string
  resourcePattern: string
  conditions?: Condition[]
}

const effectOptions = [
  { name: 'Allow', value: 'allow' },
  { name: 'Deny', value: 'deny' },
]

const conditionOperatorOptions = [
  { name: 'StringEquals', value: 'StringEquals' },
  { name: 'StringLike', value: 'StringLike' },
  { name: 'StringNotEquals', value: 'StringNotEquals' },
  { name: 'StringNotLike', value: 'StringNotLike' },
  { name: 'Bool', value: 'Bool' },
  { name: 'NumericEquals', value: 'NumericEquals' },
  { name: 'NumericGreaterThanEquals', value: 'NumericGreaterThanEquals' },
  { name: 'NumericLessThanEquals', value: 'NumericLessThanEquals' },
  { name: 'DateGreaterThan', value: 'DateGreaterThan' },
  { name: 'DateLessThan', value: 'DateLessThan' },
  { name: 'IpAddress', value: 'IpAddress' },
  { name: 'Null', value: 'Null' },
]

function normalizeStatements(raw: string): Statement[] {
  try {
    const parsed = JSON.parse(raw || '[]')
    if (!Array.isArray(parsed)) return []
    return parsed.map((item) => ({
      effect: typeof item?.effect === 'string' ? item.effect : 'allow',
      actionPattern:
        typeof item?.actionPattern === 'string' ? item.actionPattern : '',
      resourcePattern:
        typeof item?.resourcePattern === 'string' ? item.resourcePattern : '*',
      conditions: Array.isArray(item?.conditions)
        ? item.conditions
            .map((condition: unknown) => {
              const current = condition as Partial<Condition> | null
              return {
                operator:
                  typeof current?.operator === 'string'
                    ? current.operator
                    : 'StringEquals',
                key: typeof current?.key === 'string' ? current.key : '',
                value: typeof current?.value === 'string' ? current.value : '',
              }
            })
            .filter((condition: Condition) => condition.key.trim() !== '')
        : [],
    }))
  } catch {
    return []
  }
}

function serializeStatements(items: Statement[]) {
  return JSON.stringify(items, null, 2)
}

export function IamStatementBuilder(props: {
  label: string
  value: string
  onChange: (value: string) => void
  actionOptions?: SelectOption[]
  builderCopy?: string
}) {
  const [mode, setMode] = createSignal<'builder' | 'json'>('builder')
  const [statements, setStatements] = createSignal<Statement[]>(
    normalizeStatements(props.value)
  )
  const [parseError, setParseError] = createSignal('')

  createEffect(() => {
    try {
      const next = normalizeStatements(props.value)
      setStatements(next)
      setParseError('')
    } catch {
      setParseError('Invalid JSON')
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
    return props.actionOptions?.some((option) => option.value === value)
      ? value
      : '__custom__'
  }

  const commit = (next: Statement[]) => {
    setStatements(next)
    props.onChange(serializeStatements(next))
  }

  const updateStatement = (index: number, patch: Partial<Statement>) => {
    const next = statements().map((item, currentIndex) =>
      currentIndex === index ? { ...item, ...patch } : item
    )
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

  const addCondition = (index: number) => {
    const current = statements()[index]
    updateStatement(index, {
      conditions: [
        ...(current.conditions || []),
        { operator: 'StringEquals', key: '', value: '' },
      ],
    })
  }

  const updateCondition = (
    statementIndex: number,
    conditionIndex: number,
    patch: Partial<Condition>
  ) => {
    const current = statements()[statementIndex]
    const nextConditions = (current.conditions || []).map(
      (item, currentIndex) =>
        currentIndex === conditionIndex ? { ...item, ...patch } : item
    )
    updateStatement(statementIndex, { conditions: nextConditions })
  }

  const removeCondition = (statementIndex: number, conditionIndex: number) => {
    const current = statements()[statementIndex]
    updateStatement(statementIndex, {
      conditions: (current.conditions || []).filter(
        (_, currentIndex) => currentIndex !== conditionIndex
      ),
    })
  }

  const handleJsonInput = (value: string) => {
    props.onChange(value)
    try {
      const next = normalizeStatements(value)
      setStatements(next)
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
            {props.builderCopy ||
              'Build statements with structured fields, or switch to raw JSON for advanced edits.'}
          </p>
        </div>
        <Badge content={`${statementCount()} statements`} color="blue" />
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
                    <Badge
                      content={`Statement ${statementIndex() + 1}`}
                      color="dark"
                    />
                    <Badge
                      content={statement.effect}
                      color={statement.effect === 'deny' ? 'red' : 'green'}
                    />
                  </div>
                  <Button
                    size="xs"
                    color="red"
                    onClick={() => removeStatement(statementIndex())}
                  >
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
                    <Show
                      when={
                        actionSelectValue(statement.actionPattern) ===
                        '__custom__'
                      }
                    >
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
                    <p class="text-sm font-semibold text-gray-900">
                      Conditions
                    </p>
                    <Button
                      size="xs"
                      color="light"
                      onClick={() => addCondition(statementIndex())}
                    >
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
                                updateCondition(
                                  statementIndex(),
                                  conditionIndex(),
                                  {
                                    operator: event.currentTarget.value,
                                  }
                                )
                              }
                            />
                            <InputField
                              label="Key"
                              value={condition.key}
                              placeholder="principal_tag:team"
                              onInput={(event) =>
                                updateCondition(
                                  statementIndex(),
                                  conditionIndex(),
                                  {
                                    key: event.currentTarget.value,
                                  }
                                )
                              }
                            />
                            <InputField
                              label="Value"
                              value={condition.value}
                              placeholder="ops"
                              onInput={(event) =>
                                updateCondition(
                                  statementIndex(),
                                  conditionIndex(),
                                  {
                                    value: event.currentTarget.value,
                                  }
                                )
                              }
                            />
                            <Button
                              size="xs"
                              color="red"
                              onClick={() =>
                                removeCondition(
                                  statementIndex(),
                                  conditionIndex()
                                )
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
