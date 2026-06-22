import { createContext, useContext } from 'solid-js';

// Temporary local view-model context while the legacy settings page is split.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type AdminSettingsViewModel = any;

export const AdminSettingsContext = createContext<AdminSettingsViewModel>();

export function useAdminSettings() {
  const value = useContext(AdminSettingsContext);
  if (!value) {
    throw new Error('AdminSettingsContext is missing');
  }
  return value;
}
