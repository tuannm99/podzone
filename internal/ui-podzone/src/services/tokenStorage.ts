const TOKEN_KEY = 'access_token'
const USER_KEY = 'user_info'

export type StoredUser = {
  id?: string
  username?: string
  email?: string
  [key: string]: unknown
}

function parseUser(raw: string | null): StoredUser | null {
  if (!raw) return null

  try {
    return JSON.parse(raw) as StoredUser
  } catch {
    return null
  }
}

export const tokenStorage = {
  getToken(): string {
    return localStorage.getItem(TOKEN_KEY) || ''
  },
  setToken(token: string): void {
    if (!token) return
    localStorage.setItem(TOKEN_KEY, token)
  },
  clearToken(): void {
    localStorage.removeItem(TOKEN_KEY)
  },

  getUser(): StoredUser | null {
    return parseUser(localStorage.getItem(USER_KEY))
  },
  setUser(user: StoredUser): void {
    localStorage.setItem(USER_KEY, JSON.stringify(user))
  },
  clearUser(): void {
    localStorage.removeItem(USER_KEY)
  },

  clearAll(): void {
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(USER_KEY)
  }
}
