import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';

@Component({
  selector: 'app-kbd',
  templateUrl: './kbd.component.html',
})
export class Kbd {
  class = input<string>();

  protected className = computed(() =>
    classes(
      'inline-flex min-h-7 items-center rounded-lg border border-gray-200 bg-white px-2.5 text-xs font-semibold uppercase tracking-wide text-gray-600 shadow-sm',
      this.class(),
    ),
  );
}
