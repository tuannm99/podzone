import { Component, computed, input, signal } from '@angular/core';
import { classes } from '../../utils';
import { type AlertTone, toneClasses } from '../display-shared';

@Component({
  selector: 'app-banner',
  templateUrl: './banner.component.html',
})
export class Banner {
  tone = input<AlertTone>('dark');
  title = input<string>();
  dismissible = input(false);
  class = input<string>();

  protected dismissed = signal(false);

  protected wrapperClass = computed(() =>
    classes(
      'flex flex-col gap-3 rounded-lg border px-4 py-3 shadow-sm md:flex-row md:items-center md:justify-between',
      toneClasses[this.tone()],
      this.class(),
    ),
  );
}
