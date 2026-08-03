import { Component, computed, inject, signal } from '@angular/core';
import { MatIconModule } from '@angular/material/icon';
import { RouterLink, RouterOutlet } from '@angular/router';

import { AuthService } from '../../../core/auth/auth.service';
import { Button } from '../../../shared/ui/button/button.component';
import { NavList, type NavListItem } from '../../../shared/ui/nav-list/nav-list.component';

const NAV_ITEMS: NavListItem[] = [
  { label: 'Onboarding review', href: '/admin/onboarding', exact: false },
  { label: 'Organizations', href: '/admin/iam/organizations', exact: false },
  { label: 'Platform roles', href: '/admin/iam/platform-roles', exact: false },
  { label: 'Policies', href: '/admin/iam/policies', exact: false },
];

// Dedicated top-level layout for /admin/* rather than reusing Shell
// verbatim — Shell's workspace-switcher chrome and nav items are
// tenant-scoped (PZEP-0006/0011), but /admin is a sibling of `t/:tenantId`,
// not a tenant context at all (see PZEP-0020's Open Questions). Mirrors
// Shell's structure/visual language for consistency with the rest of the
// app.
@Component({
  selector: 'app-admin-shell',
  imports: [RouterOutlet, RouterLink, NavList, MatIconModule, Button],
  templateUrl: './admin-shell.component.html',
  styleUrl: './admin-shell.component.scss',
})
export class AdminShell {
  private readonly auth = inject(AuthService);

  protected readonly navItems = computed<NavListItem[]>(() => NAV_ITEMS);

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
