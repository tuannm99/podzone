import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';

@Component({
  selector: 'app-inline-code',
  templateUrl: './inline-code.component.html',
})
export class InlineCode {
  class = input<string>();

  protected className = computed(() =>
    classes('rounded-md bg-gray-100 px-1.5 py-0.5 text-sm text-gray-800', this.class()),
  );
}
