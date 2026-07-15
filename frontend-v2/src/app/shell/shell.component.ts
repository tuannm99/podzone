import { Component } from '@angular/core';
import { MatIconModule } from '@angular/material/icon';
import { RouterLink, RouterOutlet } from '@angular/router';
import { NavList, type NavListItem } from '../shared/ui/nav-list/nav-list.component';

@Component({
  selector: 'app-shell',
  imports: [RouterOutlet, RouterLink, NavList, MatIconModule],
  templateUrl: './shell.component.html',
  styleUrl: './shell.component.scss',
})
export class Shell {
  protected readonly platformNavItems: NavListItem[] = [{ label: 'Home', href: '/', exact: true }];
}
