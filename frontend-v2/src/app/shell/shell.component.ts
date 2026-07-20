import { Component, computed, inject, signal } from '@angular/core';
import { MatIconModule } from '@angular/material/icon';
import { RouterLink, RouterOutlet } from '@angular/router';
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

  protected readonly platformNavItems: NavListItem[] = [{ label: 'Home', href: '/', exact: true }];

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
