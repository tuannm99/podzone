import { HttpContextToken, HttpErrorResponse, type HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { catchError, from, switchMap, throwError } from 'rxjs';

import { environment } from '../../../environments/environment';
import { TokenStorageService, type StoredUser } from '../storage/token-storage.service';

const REFRESH_EXEMPT_PATHS = ['/auth/v1/login', '/auth/v1/register', '/auth/v1/refresh'];
const RETRIED = new HttpContextToken<boolean>(() => false);

type RefreshResponseData = {
  jwtToken?: string;
  refreshToken?: string;
  userInfo?: unknown;
};

let refreshInFlight: Promise<string | null> | null = null;

async function performRefresh(tokenStorage: TokenStorageService): Promise<string | null> {
  const refreshToken = tokenStorage.getRefreshToken();
  if (!refreshToken) return null;

  try {
    const response = await fetch(`${environment.apiBaseUrl}/auth/v1/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refreshToken }),
    });
    if (!response.ok) return null;

    const data = (await response.json()) as RefreshResponseData;
    if (data.jwtToken) tokenStorage.setToken(data.jwtToken);
    if (data.refreshToken) tokenStorage.setRefreshToken(data.refreshToken);
    if (data.userInfo && typeof data.userInfo === 'object') {
      tokenStorage.setUser(data.userInfo as StoredUser);
    }
    return data.jwtToken || null;
  } catch {
    return null;
  }
}

function refreshAccessToken(tokenStorage: TokenStorageService): Promise<string | null> {
  if (!refreshInFlight) {
    refreshInFlight = performRefresh(tokenStorage).finally(() => {
      refreshInFlight = null;
    });
  }
  return refreshInFlight;
}

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const tokenStorage = inject(TokenStorageService);
  const router = inject(Router);

  const token = tokenStorage.getToken();
  const authedReq = token ? req.clone({ setHeaders: { Authorization: `Bearer ${token}` } }) : req;

  return next(authedReq).pipe(
    catchError((error: unknown) => {
      const isUnauthorized = error instanceof HttpErrorResponse && error.status === 401;
      const alreadyRetried = req.context.get(RETRIED);
      const isExempt = REFRESH_EXEMPT_PATHS.some((path) => req.url.includes(path));

      if (!isUnauthorized || alreadyRetried || isExempt) {
        return throwError(() => error);
      }

      return from(refreshAccessToken(tokenStorage)).pipe(
        switchMap((nextToken) => {
          if (!nextToken) {
            tokenStorage.clearAll();
            void router.navigateByUrl('/login');
            return throwError(() => error);
          }
          const retriedReq = req.clone({
            setHeaders: { Authorization: `Bearer ${nextToken}` },
            context: req.context.set(RETRIED, true),
          });
          return next(retriedReq);
        }),
      );
    }),
  );
};
