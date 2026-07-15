import { Component, input } from '@angular/core';
import { NavLink } from '../nav-link/nav-link.component';

export type NavListItem = {
  label: string;
  href: string;
  exact?: boolean;
  tag?: string;
};

@Component({
  selector: 'app-nav-list',
  imports: [NavLink],
  templateUrl: './nav-list.component.html',
  styleUrl: './nav-list.component.scss',
})
export class NavList {
  label = input<string>();
  items = input.required<NavListItem[]>();
}
