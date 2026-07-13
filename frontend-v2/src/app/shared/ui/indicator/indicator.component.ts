import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';
import { indicatorClasses } from '../display-shared';

export type IndicatorColor = 'blue' | 'green' | 'yellow' | 'red' | 'gray';

@Component({
  selector: 'app-indicator',
  templateUrl: './indicator.component.html',
})
export class Indicator {
  color = input<IndicatorColor>('green');
  ping = input(false);
  class = input<string>();

  protected wrapperClass = computed(() => classes('relative inline-flex size-3', this.class()));

  protected pingClass = computed(() =>
    classes(
      'absolute inline-flex h-full w-full animate-ping rounded-full opacity-75',
      indicatorClasses[this.color()],
    ),
  );

  protected dotClass = computed(() =>
    classes('relative inline-flex size-3 rounded-full', indicatorClasses[this.color()]),
  );
}
