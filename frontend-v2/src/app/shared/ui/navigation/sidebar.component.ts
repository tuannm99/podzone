import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';
import { NavAction } from './nav-action.component';
import type { NavItem } from './nav-item';

@Component({
  selector: 'app-sidebar',
  imports: [NavAction],
  templateUrl: './sidebar.component.html',
})
export class Sidebar {
  title = input<string>();
  items = input.required<NavItem[]>();
  class = input<string>();
  // Solid's `footer` (JSX.Element) becomes a content-projection slot —
  // use <div slot="footer"> inside <app-sidebar>. Solid's `children`
  // (rendered under the title) becomes the default (unslotted) projection.

  protected className = computed(() =>
    classes('rounded-lg border border-gray-200 bg-white p-4 shadow-sm', this.class()),
  );
}
