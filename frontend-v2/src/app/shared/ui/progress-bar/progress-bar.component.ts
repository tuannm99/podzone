import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';
import { type AlertTone, toneFillClasses } from '../display-shared';

@Component({
  selector: 'app-progress-bar',
  templateUrl: './progress-bar.component.html',
})
export class ProgressBar {
  value = input.required<number>();
  max = input(100);
  label = input<string>();
  tone = input<AlertTone>('blue');
  showValue = input(false);
  class = input<string>();

  protected percent = computed(() => {
    const max = this.max();
    if (max <= 0) return 0;
    return Math.max(0, Math.min(100, Math.round((this.value() / max) * 100)));
  });

  protected wrapperClass = computed(() => classes('space-y-2', this.class()));

  protected fillClass = computed(() =>
    classes('h-full rounded-full transition-[width]', toneFillClasses[this.tone()]),
  );
}
