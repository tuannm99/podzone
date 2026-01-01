import React, { createContext, useContext, useMemo, useState } from 'react';
import { loadAuth, saveAuth, clearAuth } from './auth.storage';

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [auth, setAuth] = useState(() => loadAuth());

  const value = useMemo(
    () => ({
      auth,
      isAuthenticated: !!auth?.accessToken,
      setSession: (session) => {
        setAuth(session);
        saveAuth(session);
      },
      logout: () => {
        setAuth(null);
        clearAuth();
      },
    }),
    [auth],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function UseAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used inside AuthProvider');
  return ctx;
}
