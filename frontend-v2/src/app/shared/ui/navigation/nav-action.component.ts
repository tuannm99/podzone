import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';
// NOTE: depends on Link.tsx's Angular port, assigned to a different
// parallel porting batch. Path assumed per this project's folder-per-
// component convention — reconcile at the consolidated build pass if the
// actual path differs.
import { Link } from '../link/link.component';
import type { NavItem } from './nav-item';

@Component({
  selector: 'app-nav-action',
  imports: [Link],
  templateUrl: './nav-action.component.html',
})
export class NavAction {
  item = input.required<NavItem>();
  class = input<string>();
  activeClass = input<string>();

  protected className = computed(() =>
    classes(
      'inline-flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium transition',
      this.item().active
        ? (this.activeClass() ?? 'bg-gray-950 text-white')
        : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900',
      this.class(),
    ),
  );

  protected onClick() {
    this.item().onClick?.();
  }
}
