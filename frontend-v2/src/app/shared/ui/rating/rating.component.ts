import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';

@Component({
  selector: 'app-rating',
  templateUrl: './rating.component.html',
})
export class Rating {
  value = input.required<number>();
  max = input(5);
  class = input<string>();

  protected stars = computed(() => Array.from({ length: this.max() }, (_, index) => index + 1));

  protected wrapperClass = computed(() => classes('inline-flex items-center gap-1', this.class()));
}
