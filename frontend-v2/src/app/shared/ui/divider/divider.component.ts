import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';

@Component({
  selector: 'app-divider',
  templateUrl: './divider.component.html',
})
export class Divider {
  label = input<string>();
  class = input<string>();

  protected className = computed(() => classes('flex items-center gap-4', this.class()));
}
