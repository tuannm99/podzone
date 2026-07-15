import { HttpClient } from '@angular/common/http';
import { Injectable, computed, inject, signal } from '@angular/core';
import { Router } from '@angular/router';
import { firstValueFrom } from 'rxjs';

import { environment } from '../../../environments/environment';
import { httpErrorMessage } from '../http/http-error';
import { TenantStorageService } from '../storage/tenant-storage.service';
import { TokenStorageService, type StoredUser } from '../storage/token-storage.service';

export type AuthResponseData = {
  jwtToken?: string;
  refreshToken?: string;
  userInfo?: StoredUser;
  message?: string;
  [key: string]: unknown;
};

export type AuthResult =
  | { success: true; data: AuthResponseData; message: string }
  | { success: false; data: null; message: string };

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly http = inject(HttpClient);
  private readonly tokenStorage = inject(TokenStorageService);
  private readonly tenantStorage = inject(TenantStorageService);
  private readonly router = inject(Router);

  private readonly _currentUser = signal<StoredUser | null>(this.tokenStorage.getUser());
  readonly currentUser = this._currentUser.asReadonly();
  readonly isAuthenticated = computed(() => this._currentUser() !== null);

  async login(payload: { username: string; password: string }): Promise<AuthResult> {
    try {
      const data = await firstValueFrom(
        this.http.post<AuthResponseData>(`${environment.apiBaseUrl}/auth/v1/login`, payload),
      );
      this.persistAuth(data);
      return { success: true, data, message: '' };
    } catch (error) {
      return this.toFailure(error, 'Login failed');
    }
  }

  async register(payload: {
    username: string;
    email: string;
    password: string;
  }): Promise<AuthResult> {
    try {
      const data = await firstValueFrom(
        this.http.post<AuthResponseData>(`${environment.apiBaseUrl}/auth/v1/register`, payload),
      );
      this.persistAuth(data);
      return { success: true, data, message: '' };
    } catch (error) {
      return this.toFailure(error, 'Register failed');
    }
  }

  googleLoginUrl(): string {
    return `${environment.apiBaseUrl}/auth/v1/google/login`;
  }

  async ensureActiveTenant(tenantId: string): Promise<AuthResult> {
    const normalized = tenantId.trim();
    if (!normalized) {
      return { success: false, data: null, message: 'Tenant id is required' };
    }

    if (this.tokenStorage.getActiveTenantId() === normalized) {
      this.tenantStorage.setTenantId(normalized);
      return {
        success: true,
        data: {
          jwtToken: this.tokenStorage.getToken(),
          userInfo: this.tokenStorage.getUser() ?? undefined,
        },
        message: '',
      };
    }

    return this.switchActiveTenant(normalized);
  }

  async switchActiveTenant(tenantId: string): Promise<AuthResult> {
    const normalized = tenantId.trim();
    if (!normalized) {
      return { success: false, data: null, message: 'Tenant id is required' };
    }

    const userId = this.tokenStorage.getUserId();
    if (!userId) {
      return { success: false, data: null, message: 'No authenticated user found' };
    }
    const accessToken = this.tokenStorage.getToken();
    if (!accessToken) {
      return { success: false, data: null, message: 'No active session token found' };
    }

    try {
      const data = await firstValueFrom(
        this.http.post<AuthResponseData>(`${environment.apiBaseUrl}/auth/v1/iam/tenants:switch`, {
          userId,
          tenantId: normalized,
          accessToken,
        }),
      );
      this.persistAuth(data);
      this.tenantStorage.setTenantId(normalized);
      return { success: true, data, message: '' };
    } catch (error) {
      return this.toFailure(error, 'Switch tenant failed');
    }
  }

  async logout(): Promise<void> {
    const token = this.tokenStorage.getToken();
    try {
      if (token) {
        await firstValueFrom(this.http.post(`${environment.apiBaseUrl}/auth/v1/logout`, { token }));
      }
    } catch {
      // Revoke on server is best-effort; local credentials are cleared regardless.
    } finally {
      this.tokenStorage.clearAll();
      this.tenantStorage.clearTenantId();
      this._currentUser.set(null);
      void this.router.navigateByUrl('/login');
    }
  }

  private persistAuth(data: AuthResponseData): void {
    if (data.jwtToken) this.tokenStorage.setToken(data.jwtToken);
    if (data.refreshToken) this.tokenStorage.setRefreshToken(data.refreshToken);
    if (data.userInfo) {
      const existingId = this.tokenStorage.getUserId();
      const user: StoredUser = {
        ...data.userInfo,
        id: data.userInfo.id ?? (existingId != null ? String(existingId) : undefined),
      };
      this.tokenStorage.setUser(user);
      this._currentUser.set(user);
    }
    const activeTenantId = this.tokenStorage.getActiveTenantId();
    if (activeTenantId) this.tenantStorage.setTenantId(activeTenantId);
  }

  private toFailure(error: unknown, fallback: string): AuthResult {
    return { success: false, data: null, message: httpErrorMessage(error, fallback) };
  }
}
