import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';

@Component({
  selector: 'app-paragraph',
  templateUrl: './paragraph.component.html',
})
export class Paragraph {
  muted = input(false);
  class = input<string>();

  protected className = computed(() =>
    classes('text-base leading-7', this.muted() ? 'text-gray-500' : 'text-gray-700', this.class()),
  );
}
