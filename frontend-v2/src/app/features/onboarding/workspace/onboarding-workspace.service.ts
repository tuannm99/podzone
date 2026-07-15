import { HttpClient } from '@angular/common/http';
import { Injectable, computed, inject, resource } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { environment } from '../../../../environments/environment';
import { AuthService } from '../../../core/auth/auth.service';
import { httpErrorMessage } from '../../../core/http/http-error';
import { TenantStorageService } from '../../../core/storage/tenant-storage.service';
import type { TenantMembership } from './workspace.types';

type ListTenantsResult =
  | { success: true; data: TenantMembership[]; message: string }
  | { success: false; data: null; message: string };

@Injectable()
export class OnboardingWorkspaceService {
  private readonly http = inject(HttpClient);
  private readonly auth = inject(AuthService);
  private readonly tenantStorage = inject(TenantStorageService);

  private readonly userId = computed(() => {
    const user = this.auth.currentUser();
    const parsed = user?.id != null ? Number.parseInt(user.id, 10) : NaN;
    return Number.isFinite(parsed) ? parsed : null;
  });

  readonly memberships = resource({
    params: () => (this.userId() != null ? { userId: this.userId()! } : undefined),
    loader: async ({ params }) => {
      const result = await this.listUserTenants(params.userId);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });

  readonly activeMemberships = computed(() =>
    (this.memberships.value() ?? []).filter((membership) => membership.status === 'active'),
  );

  readonly selectedTenantId = computed(() => this.tenantStorage.getTenantId());

  async selectWorkspace(tenantId: string): Promise<{ success: boolean; message: string }> {
    const result = await this.auth.ensureActiveTenant(tenantId);
    if (!result.success) {
      return { success: false, message: result.message };
    }
    this.tenantStorage.setTenantId(tenantId);
    return { success: true, message: '' };
  }

  private async listUserTenants(userId: number): Promise<ListTenantsResult> {
    try {
      const response = await firstValueFrom(
        this.http.get<{ memberships?: TenantMembership[] }>(
          `${environment.apiBaseUrl}/auth/v1/iam/users/${userId}/tenants`,
        ),
      );
      return { success: true, data: response.memberships || [], message: '' };
    } catch (error) {
      return {
        success: false,
        data: null,
        message: httpErrorMessage(error, 'Failed to load user tenants'),
      };
    }
  }
}
