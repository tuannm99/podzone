import { Component, computed, input, output } from '@angular/core';
import { FieldLabel } from '../field-label/field-label.component';
import { fieldBaseClasses } from '../field-classes';
import { createUniqueId } from '../../utils';

@Component({
  selector: 'app-input-field',
  imports: [FieldLabel],
  templateUrl: './input-field.component.html',
})
export class InputField {
  label = input.required<string>();
  value = input.required<string>();
  id = input<string>(createUniqueId());
  type = input('text');
  placeholder = input<string>();
  disabled = input(false);
  error = input(false);
  errorText = input<string>();

  valueChange = output<string>();

  protected fieldClass = computed(() => fieldBaseClasses(this.error()));
  protected errorId = computed(() => (this.errorText() ? `${this.id()}-error` : null));
}
