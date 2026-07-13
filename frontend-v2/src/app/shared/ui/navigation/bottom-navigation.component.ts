import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';
import { NavAction } from './nav-action.component';
import type { NavItem } from './nav-item';

@Component({
  selector: 'app-bottom-navigation',
  imports: [NavAction],
  templateUrl: './bottom-navigation.component.html',
})
export class BottomNavigation {
  items = input.required<NavItem[]>();
  class = input<string>();

  protected className = computed(() =>
    classes(
      'fixed inset-x-0 bottom-0 z-40 border-t border-gray-200 bg-white/95 px-4 py-2 shadow-[0_-8px_24px_rgba(15,23,42,0.08)] backdrop-blur md:hidden',
      this.class(),
    ),
  );
}
