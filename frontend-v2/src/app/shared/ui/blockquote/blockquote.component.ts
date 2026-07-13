import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';

@Component({
  selector: 'app-blockquote',
  templateUrl: './blockquote.component.html',
})
export class Blockquote {
  cite = input<string>();
  class = input<string>();

  protected className = computed(() =>
    classes('rounded-lg border-s-4 border-gray-900 bg-gray-50 px-5 py-4', this.class()),
  );
}
