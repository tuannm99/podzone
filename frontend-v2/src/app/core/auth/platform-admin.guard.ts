import { HttpClient } from '@angular/common/http';
import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { firstValueFrom } from 'rxjs';

import { environment } from '../../../environments/environment';
import { TokenStorageService } from '../storage/token-storage.service';

type CheckPlatformPermissionResponse = { allowed?: boolean };

// Gates every /admin/* route (PZEP-0020). Calls the legacy IAMService's
// CheckPlatformPermission RPC (already REST-mapped, already used
// server-side by AuthorizeStoreApproval/AuthorizeStoreReadAll) — this is
// the first frontend consumer of a platform role for anything. Fails
// closed: no session, no user id, a request error, or `allowed: false` all
// deny — matches SRS-NFR-002. This guard alone is not a security boundary
// (per this repo's own rule); the interactor methods it gates re-check the
// same permission server-side regardless of what this guard decides.
export function platformAdminGuard(permission: string): CanActivateFn {
  return async () => {
    const tokenStorage = inject(TokenStorageService);
    const http = inject(HttpClient);
    const router = inject(Router);

    const userId = tokenStorage.getUserId();
    if (userId == null) {
      return router.parseUrl('/login');
    }

    try {
      const response = await firstValueFrom(
        http.post<CheckPlatformPermissionResponse>(
          `${environment.apiBaseUrl}/auth/v1/iam/platform-permissions:check`,
          { userId, permission },
        ),
      );
      return response.allowed === true || router.parseUrl('/');
    } catch {
      return router.parseUrl('/');
    }
  };
}
