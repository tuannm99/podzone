import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';
import { type AlertTone, toneClasses } from '../display-shared';

export type { AlertTone };

@Component({
  selector: 'app-alert',
  templateUrl: './alert.component.html',
})
export class Alert {
  tone = input<AlertTone>('blue');
  title = input<string>();
  class = input<string>();

  protected wrapperClass = computed(() =>
    classes(
      'flex flex-col gap-3 rounded-lg border px-4 py-3 shadow-sm md:flex-row md:items-start md:justify-between',
      toneClasses[this.tone()],
      this.class(),
    ),
  );
}
