import { Component, computed, inject, signal } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { MatIconModule } from '@angular/material/icon';
import { ActivatedRoute, NavigationEnd, Router, RouterLink, RouterOutlet } from '@angular/router';
import { filter, map, startWith } from 'rxjs';
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
  private readonly router = inject(Router);

  private readonly paramMap = toSignal(this.route.paramMap, { requireSync: true });
  private readonly tenantId = computed(() => this.paramMap().get('tenantId') ?? '');

  protected readonly platformNavItems = computed<NavListItem[]>(() => {
    const tenantId = this.tenantId();
    return [
      { label: 'Home', href: `/t/${tenantId}/backoffice`, exact: true },
      { label: 'Stores', href: `/t/${tenantId}/backoffice/stores`, exact: true },
      { label: 'Orders', href: `/t/${tenantId}/backoffice/orders`, exact: true },
    ];
  });

  // Reflects the active nav item in the header instead of PZEP-0006's
  // hardcoded "Choose a store" placeholder, which read as wrong/confusing
  // once real Backoffice content (past the handoff) started rendering here.
  private readonly currentUrl = toSignal(
    this.router.events.pipe(
      filter((event) => event instanceof NavigationEnd),
      map(() => this.router.url),
      startWith(this.router.url),
    ),
    { requireSync: true },
  );

  protected readonly headerTitle = computed(() => {
    const url = this.currentUrl();
    const active = this.platformNavItems().find((item) => url.startsWith(item.href));
    return active?.label ?? 'Backoffice';
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
