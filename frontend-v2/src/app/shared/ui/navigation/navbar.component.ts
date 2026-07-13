import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';
import { NavAction } from './nav-action.component';
import type { NavItem } from './nav-item';

@Component({
  selector: 'app-navbar',
  imports: [NavAction],
  templateUrl: './navbar.component.html',
})
export class Navbar {
  items = input<NavItem[]>([]);
  class = input<string>();
  // Solid's `brand`/`actions` (JSX.Element props) become content-
  // projection slots — use <div slot="brand"> / <div slot="actions">
  // inside <app-navbar>.

  protected className = computed(() =>
    classes('rounded-lg border border-gray-200 bg-white px-4 py-3 shadow-sm', this.class()),
  );
}
