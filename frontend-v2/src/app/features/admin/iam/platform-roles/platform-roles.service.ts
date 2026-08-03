import { HttpClient } from '@angular/common/http';
import { Injectable, inject, resource, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { environment } from '../../../../../environments/environment';
import { httpErrorMessage } from '../../../../core/http/http-error';
import {
  toCollectionParams,
  type CollectionQuery,
} from '../../../../shared/services/collection-types';
import type { PlatformRoleMembership } from '../iam.types';

type ActionResult = { success: boolean; message: string };

const QUERY: CollectionQuery = { page: 1, pageSize: 50, sortBy: 'createdAt' };

// Route-scoped: owns the platform-role-assignment lookup for a given user
// (PZEP-0020) - the same operation cmd/iam-bootstrap (PZEP-0002) performs
// from the CLI, now with a UI. Ported from frontend/packages/shared/
// services/iam/roles.ts's listPlatformRoles/upsertPlatformRole/
// removePlatformRole.
@Injectable()
export class PlatformRolesService {
  private readonly http = inject(HttpClient);

  private readonly _targetUserId = signal<number | null>(null);
  private readonly _mutating = signal(false);

  readonly targetUserId = this._targetUserId.asReadonly();
  readonly mutating = this._mutating.asReadonly();

  readonly roles = resource({
    params: () =>
      this._targetUserId() != null ? { targetUserId: this._targetUserId()! } : undefined,
    loader: async ({ params }) => {
      try {
        const response = await firstValueFrom(
          this.http.get<{ memberships?: PlatformRoleMembership[] }>(
            `${environment.apiBaseUrl}/auth/v1/iam/platform-users/${params.targetUserId}/roles`,
            { params: toCollectionParams(QUERY) },
          ),
        );
        return response.memberships ?? [];
      } catch (error) {
        throw new Error(httpErrorMessage(error, 'Failed to load platform roles'));
      }
    },
  });

  lookupUser(targetUserId: number): void {
    this._targetUserId.set(targetUserId);
  }

  async addRole(roleName: string): Promise<ActionResult> {
    const targetUserId = this._targetUserId();
    if (targetUserId == null) return { success: false, message: 'Look up a user first' };
    return this.mutate(async () => {
      await firstValueFrom(
        this.http.post(
          `${environment.apiBaseUrl}/auth/v1/iam/platform-users/${targetUserId}/roles`,
          {
            targetUserId,
            roleName,
          },
        ),
      );
      this.roles.reload();
    }, 'Failed to save platform role');
  }

  async removeRole(roleName: string): Promise<ActionResult> {
    const targetUserId = this._targetUserId();
    if (targetUserId == null) return { success: false, message: 'Look up a user first' };
    return this.mutate(async () => {
      await firstValueFrom(
        this.http.delete(
          `${environment.apiBaseUrl}/auth/v1/iam/platform-users/${targetUserId}/roles/${roleName}`,
        ),
      );
      this.roles.reload();
    }, 'Failed to remove platform role');
  }

  private async mutate(action: () => Promise<void>, fallback: string): Promise<ActionResult> {
    this._mutating.set(true);
    try {
      await action();
      return { success: true, message: '' };
    } catch (error) {
      return { success: false, message: httpErrorMessage(error, fallback) };
    } finally {
      this._mutating.set(false);
    }
  }
}
