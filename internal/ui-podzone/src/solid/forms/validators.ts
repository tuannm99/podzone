export type Validator<TValues, TKey extends keyof TValues = keyof TValues> = (
    value: TValues[TKey],
    values: TValues
) => string | undefined

export type ValidatorMap<TValues> = {
    [K in keyof TValues]?: Validator<TValues, K>[]
}

export function required<TValues, TKey extends keyof TValues>(
    message = 'This field is required.'
): Validator<TValues, TKey> {
    return (value) => {
        if (typeof value === 'string' && value.trim().length === 0) {
            return message
        }
        if (value === null || value === undefined) {
            return message
        }
        return undefined
    }
}

export function email<TValues, TKey extends keyof TValues>(
    message = 'Enter a valid email address.'
): Validator<TValues, TKey> {
    return (value) => {
        if (typeof value !== 'string' || value.trim().length === 0) {
            return undefined
        }
        return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value.trim()) ? undefined : message
    }
}

export function numberValue<TValues, TKey extends keyof TValues>(
    message = 'Enter a valid number.'
): Validator<TValues, TKey> {
    return (value) => {
        if (typeof value !== 'string' || value.trim().length === 0) {
            return undefined
        }
        return Number.isFinite(Number(value)) ? undefined : message
    }
}

export function jsonArray<TValues, TKey extends keyof TValues>(
    message = 'Enter a valid JSON array.'
): Validator<TValues, TKey> {
    return (value) => {
        if (typeof value !== 'string') {
            return message
        }
        try {
            return Array.isArray(JSON.parse(value || '[]')) ? undefined : message
        } catch {
            return message
        }
    }
}

export function jsonObject<TValues, TKey extends keyof TValues>(
    message = 'Enter a valid JSON object.'
): Validator<TValues, TKey> {
    return (value) => {
        if (typeof value !== 'string') {
            return message
        }
        try {
            const parsed = JSON.parse(value || '{}')
            return parsed && typeof parsed === 'object' && !Array.isArray(parsed) ? undefined : message
        } catch {
            return message
        }
    }
}
