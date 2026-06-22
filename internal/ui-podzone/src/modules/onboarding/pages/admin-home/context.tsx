import { createContext, useContext } from 'solid-js';

// The admin home page is being split from a legacy single-file page. Keep this
// context local to the page folder while the controller is extracted next.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type AdminHomeViewModel = any;

export const AdminHomeContext = createContext<AdminHomeViewModel>();

export function useAdminHome() {
  const value = useContext(AdminHomeContext);
  if (!value) {
    throw new Error('AdminHomeContext is missing');
  }
  return value;
}
