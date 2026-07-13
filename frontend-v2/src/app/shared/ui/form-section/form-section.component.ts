import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';

@Component({
  selector: 'app-form-section',
  templateUrl: './form-section.component.html',
})
export class FormSection {
  title = input.required<string>();
  description = input<string>();
  class = input<string>();
  contentClass = input<string>();

  protected className = computed(() =>
    classes('rounded-lg border border-gray-200 bg-white p-5 shadow-sm', this.class()),
  );
  protected contentClassName = computed(() => classes('mt-6 space-y-5', this.contentClass()));
}
