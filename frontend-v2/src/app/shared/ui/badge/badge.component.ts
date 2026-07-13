import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';

export type BadgeColor = 'blue' | 'indigo' | 'green' | 'yellow' | 'pink' | 'dark' | 'red';

const badgeColorClasses: Record<BadgeColor, string> = {
  blue: 'bg-blue-100 text-blue-800',
  indigo: 'bg-indigo-100 text-indigo-800',
  green: 'bg-green-100 text-green-800',
  yellow: 'bg-yellow-100 text-yellow-800',
  pink: 'bg-pink-100 text-pink-800',
  dark: 'bg-gray-100 text-gray-800',
  red: 'bg-red-100 text-red-800',
};

@Component({
  selector: 'app-badge',
  templateUrl: './badge.component.html',
})
export class Badge {
  content = input.required<string>();
  color = input<BadgeColor>('dark');
  class = input<string>();

  protected className = computed(() =>
    classes(
      'inline-flex items-center rounded-md px-2 py-1 text-xs font-semibold',
      badgeColorClasses[this.color()],
      this.class(),
    ),
  );
}
