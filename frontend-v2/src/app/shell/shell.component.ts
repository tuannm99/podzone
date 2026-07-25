import { Component, computed, inject, signal } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { MatIconModule } from '@angular/material/icon';
import { ActivatedRoute, RouterLink, RouterOutlet } from '@angular/router';
import { AuthService } from '../core/auth/auth.service';
import { Button } from '../shared/ui/button/button.component';
import { NavList, type NavListItem } from '../shared/ui/nav-list/nav-list.component';

@Component({
  selector: 'app-shell',
  imports: [RouterOutlet, RouterLink, NavList, MatIconModule, Button],
  templateUrl: './shell.component.html',
  styleUrl: './shell.component.scss',
})
export class Shell {
  private readonly auth = inject(AuthService);
  private readonly route = inject(ActivatedRoute);

  private readonly paramMap = toSignal(this.route.paramMap, { requireSync: true });
  private readonly tenantId = computed(() => this.paramMap().get('tenantId') ?? '');

  protected readonly platformNavItems = computed<NavListItem[]>(() => {
    const tenantId = this.tenantId();
    return [
      { label: 'Home', href: `/t/${tenantId}/backoffice`, exact: true },
      { label: 'Stores', href: `/t/${tenantId}/backoffice/stores`, exact: true },
    ];
  });

  protected readonly currentUser = this.auth.currentUser;
  protected readonly userInitial = computed(() => {
    const user = this.currentUser();
    const label = user?.username || user?.email || '';
    return label ? label.charAt(0).toUpperCase() : '?';
  });

  protected logout() {
    void this.auth.logout();
  }

  protected readonly navCollapsed = signal(false);

  protected toggleNav() {
    this.navCollapsed.update((collapsed) => !collapsed);
  }
}
