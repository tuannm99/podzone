import { Component, computed, input, output } from '@angular/core';
import { FieldLabel } from '../field-label/field-label.component';
import { fieldBaseClasses } from '../field-classes';
import { createUniqueId } from '../../utils';

export type SelectOption = {
  name: string;
  value: string;
};

@Component({
  selector: 'app-select-field',
  imports: [FieldLabel],
  templateUrl: './select-field.component.html',
})
export class SelectField {
  label = input.required<string>();
  value = input.required<string>();
  options = input.required<SelectOption[]>();
  id = input<string>(createUniqueId());
  disabled = input(false);
  error = input(false);
  errorText = input<string>();

  valueChange = output<string>();

  protected fieldClass = computed(() => fieldBaseClasses(this.error()));
  protected errorId = computed(() => (this.errorText() ? `${this.id()}-error` : null));
}
