import { http, type HttpError } from '../http'
import { toFailure } from './result'
import type {
  IamResult,
  SimulateAccessResult,
  SimulateAccessPayload,
} from './types'

export async function simulateAccess(
  payload: SimulateAccessPayload
): Promise<IamResult<SimulateAccessResult>> {
  try {
    const { data } = await http.post<SimulateAccessResult>(
      '/auth/v1/iam/access:simulate',
      payload
    )
    return { success: true, data }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to simulate access')
  }
}
