import { http, type HttpError } from '../http'
import { toFailure } from './result'
import type {
  IamResult,
  SimulateAccessResult,
  SimulateAccessPayload,
  CheckPermissionPayload,
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

export async function checkPermission(
  payload: CheckPermissionPayload
): Promise<IamResult<boolean>> {
  try {
    const { data } = await http.post<{ allowed?: boolean }>(
      '/auth/v1/iam/permissions:check',
      payload
    )
    return { success: true, data: Boolean(data.allowed) }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to check permission')
  }
}

export async function checkPlatformPermission(
  permission: string
): Promise<IamResult<boolean>> {
  try {
    const { data } = await http.post<{ allowed?: boolean }>(
      '/auth/v1/iam/platform-permissions:check',
      { permission }
    )
    return { success: true, data: Boolean(data.allowed) }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to check platform permission')
  }
}
