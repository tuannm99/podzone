import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';

@Component({
  selector: 'app-skeleton',
  templateUrl: './skeleton.component.html',
})
export class Skeleton {
  class = input<string>();
  circle = input(false);

  protected className = computed(() =>
    classes(
      'animate-pulse bg-gray-200',
      this.circle() ? 'rounded-full' : 'rounded-md',
      this.class() ?? 'h-4 w-full',
    ),
  );
}
