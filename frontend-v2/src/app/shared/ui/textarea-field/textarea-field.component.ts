import { Component, computed, input, output } from '@angular/core';
import { FieldLabel } from '../field-label/field-label.component';
import { fieldBaseClasses } from '../field-classes';
import { classes, createUniqueId } from '../../utils';

@Component({
  selector: 'app-textarea-field',
  imports: [FieldLabel],
  templateUrl: './textarea-field.component.html',
})
export class TextareaField {
  label = input.required<string>();
  value = input.required<string>();
  id = input<string>(createUniqueId());
  rows = input(6);
  error = input(false);
  errorText = input<string>();

  valueChange = output<string>();

  protected fieldClass = computed(() => classes(fieldBaseClasses(this.error()), 'h-auto py-2.5'));
  protected errorId = computed(() => (this.errorText() ? `${this.id()}-error` : null));
}
