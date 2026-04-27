import { GW_API_URL } from './baseurl'
import { http, type HttpError } from './http'
import { tokenStorage, type StoredUser } from './tokenStorage'

export type AuthPayload = {
  username: string
  password: string
}

export type RegisterPayload = {
  username: string
  email: string
  password: string
}

export type AuthResponseData = {
  jwtToken?: string
  userInfo?: StoredUser
  message?: string
  [key: string]: unknown
}

export type AuthResult =
  | { success: true; data: AuthResponseData }
  | { success: false; data: { message: string } }

function persistAuth(data: AuthResponseData) {
  if (data.jwtToken) tokenStorage.setToken(data.jwtToken)
  if (data.userInfo) tokenStorage.setUser(data.userInfo)
}

function toFailure(error: unknown, fallback: string): AuthResult {
  const message =
    typeof error === 'object' &&
    error &&
    'message' in error &&
    typeof error.message === 'string'
      ? error.message
      : fallback

  return { success: false, data: { message } }
}

export function loginGG(): string {
  return `${GW_API_URL}/auth/v1/google/login`
}

export async function login(payload: AuthPayload): Promise<AuthResult> {
  try {
    const { data } = await http.post<AuthResponseData>('/auth/v1/login', payload)
    persistAuth(data)
    return { success: true, data }
  } catch (error) {
    return toFailure(error as HttpError, 'Login failed')
  }
}

export async function register(payload: RegisterPayload): Promise<AuthResult> {
  try {
    const { data } = await http.post<AuthResponseData>('/auth/v1/register', payload)
    persistAuth(data)
    return { success: true, data }
  } catch (error) {
    return toFailure(error as HttpError, 'Register failed')
  }
}

export function logout(): void {
  tokenStorage.clearAll()
  window.location.href = '/auth/login'
}
