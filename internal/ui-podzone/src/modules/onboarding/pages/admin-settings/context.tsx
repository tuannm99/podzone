import { createContext, useContext } from 'solid-js'
import type { AdminSettingsViewModel } from '../AdminSettingsPage'

export const AdminSettingsContext = createContext<AdminSettingsViewModel>()

export function useAdminSettings() {
  const value = useContext(AdminSettingsContext)
  if (!value) {
    throw new Error('AdminSettingsContext is missing')
  }
  return value
}
