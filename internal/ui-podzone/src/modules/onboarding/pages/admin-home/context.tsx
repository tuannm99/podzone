import { createContext, useContext } from 'solid-js'
import type { AdminHomeViewModel } from '../AdminHomePage'

export const AdminHomeContext = createContext<AdminHomeViewModel>()

export function useAdminHome() {
  const value = useContext(AdminHomeContext)
  if (!value) {
    throw new Error('AdminHomeContext is missing')
  }
  return value
}
