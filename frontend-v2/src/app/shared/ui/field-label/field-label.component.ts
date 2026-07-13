import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';

@Component({
  selector: 'app-field-label',
  templateUrl: './field-label.component.html',
})
export class FieldLabel {
  label = input.required<string>();
  for = input<string>();
  class = input<string>();

  protected className = computed(() => classes('space-y-1.5', this.class()));
}
