import { Component, inject, signal } from '@angular/core';

import { Button } from '../../../../shared/ui/button/button.component';
import { Card } from '../../../../shared/ui/card/card.component';
import { EmptyBlock } from '../../../../shared/ui/empty-block/empty-block.component';
import { ErrorAlert } from '../../../../shared/ui/error-alert/error-alert.component';
import { InputField } from '../../../../shared/ui/input-field/input-field.component';
import { LoadingBlock } from '../../../../shared/ui/loading-block/loading-block.component';
import type { PlatformRoleMembership } from '../iam.types';
import { PlatformRolesService } from './platform-roles.service';

// Platform Roles section of /admin/iam (PZEP-0020) — look up a user by id
// and manage their platform role grants (platform_owner/platform_admin),
// the same operation cmd/iam-bootstrap performs from the CLI (PZEP-0002),
// now with a UI instead of shell access.
@Component({
  selector: 'app-platform-roles-page',
  imports: [Card, Button, EmptyBlock, ErrorAlert, LoadingBlock, InputField],
  providers: [PlatformRolesService],
  templateUrl: './platform-roles-page.component.html',
  styleUrl: './platform-roles-page.component.scss',
})
export class PlatformRolesPage {
  protected readonly service = inject(PlatformRolesService);

  protected readonly userIdInput = signal('');
  protected readonly newRoleName = signal('');
  protected readonly actionError = signal('');

  protected onLookup(event: SubmitEvent) {
    event.preventDefault();
    const userId = Number(this.userIdInput());
    if (!userId) return;
    this.service.lookupUser(userId);
  }

  protected async onAddRole(event: SubmitEvent) {
    event.preventDefault();
    const roleName = this.newRoleName().trim();
    if (!roleName) return;

    const result = await this.service.addRole(roleName);
    if (result.success) {
      this.newRoleName.set('');
      this.actionError.set('');
    } else {
      this.actionError.set(result.message);
    }
  }

  protected async onRemoveRole(membership: PlatformRoleMembership) {
    const result = await this.service.removeRole(membership.roleName);
    if (!result.success) this.actionError.set(result.message);
  }
}
