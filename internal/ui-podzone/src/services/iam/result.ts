import type { IamResult } from './types'

export function toFailure(error: unknown, fallback: string): IamResult<never> {
    const message =
        typeof error === 'object' && error && 'message' in error && typeof error.message === 'string'
            ? error.message
            : fallback
    return { success: false, message }
}
