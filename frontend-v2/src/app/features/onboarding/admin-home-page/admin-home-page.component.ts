import { Component, inject, signal } from '@angular/core';
import { MatIconModule } from '@angular/material/icon';
import { Router } from '@angular/router';

import { AuthService } from '../../../core/auth/auth.service';
import { Button } from '../../../shared/ui/button/button.component';
import { Card } from '../../../shared/ui/card/card.component';
import { ToasterService } from '../../../shared/services/toaster.service';
import { OnboardingWorkspaceService } from '../workspace/onboarding-workspace.service';
import { WorkspaceChooser } from '../workspace/workspace-chooser/workspace-chooser.component';

@Component({
  selector: 'app-admin-home-page',
  imports: [Card, WorkspaceChooser, Button, MatIconModule],
  providers: [OnboardingWorkspaceService],
  templateUrl: './admin-home-page.component.html',
  styleUrl: './admin-home-page.component.scss',
})
export class AdminHomePage {
  protected readonly workspace = inject(OnboardingWorkspaceService);
  private readonly toaster = inject(ToasterService);
  private readonly router = inject(Router);
  private readonly auth = inject(AuthService);

  protected selectingTenantId = signal('');

  protected logout() {
    void this.auth.logout();
  }

  async onSelectWorkspace(tenantId: string) {
    this.selectingTenantId.set(tenantId);
    try {
      const result = await this.workspace.selectWorkspace(tenantId);
      if (!result.success) {
        this.toaster.error(result.message);
        return;
      }
      await this.router.navigate(['/t', tenantId, 'stores']);
    } finally {
      this.selectingTenantId.set('');
    }
  }
}
