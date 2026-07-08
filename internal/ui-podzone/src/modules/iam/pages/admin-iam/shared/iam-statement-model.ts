export type IamCondition = {
    operator: string
    key: string
    value: string
}

export type IamStatement = {
    effect: string
    actionPattern: string
    resourcePattern: string
    conditions?: IamCondition[]
}

export const effectOptions = [
    { name: 'Allow', value: 'allow' },
    { name: 'Deny', value: 'deny' },
]

export const conditionOperatorOptions = [
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

export function parseStatements(raw: string): IamStatement[] {
    const parsed: unknown = JSON.parse(raw || '[]')
    if (!Array.isArray(parsed)) {
        throw new Error('Statements JSON must be an array')
    }

    return parsed.map((item) => {
        const current = item as Partial<IamStatement> | null
        if (!current || typeof current !== 'object') {
            throw new Error('Each statement must be an object')
        }
        return {
            effect: typeof current.effect === 'string' ? current.effect : 'allow',
            actionPattern: typeof current.actionPattern === 'string' ? current.actionPattern : '',
            resourcePattern: typeof current.resourcePattern === 'string' ? current.resourcePattern : '*',
            conditions: parseConditions(current.conditions),
        }
    })
}

export function serializeStatements(items: IamStatement[]) {
    return JSON.stringify(items, null, 2)
}

function parseConditions(raw: unknown): IamCondition[] {
    if (!Array.isArray(raw)) return []

    return raw
        .map((condition: unknown) => {
            const value = condition as Partial<IamCondition> | null
            return {
                operator: typeof value?.operator === 'string' ? value.operator : 'StringEquals',
                key: typeof value?.key === 'string' ? value.key : '',
                value: typeof value?.value === 'string' ? value.value : '',
            }
        })
        .filter((condition) => condition.key.trim() !== '')
}
