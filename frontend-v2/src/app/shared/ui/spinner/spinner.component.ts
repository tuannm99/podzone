import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';

@Component({
  selector: 'app-spinner',
  templateUrl: './spinner.component.html',
})
export class Spinner {
  class = input<string>();

  protected className = computed(() =>
    classes(
      'inline-block size-4 animate-spin rounded-full border-2 border-current border-r-transparent',
      this.class(),
    ),
  );
}
