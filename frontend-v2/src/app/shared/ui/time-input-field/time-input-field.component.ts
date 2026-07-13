import { Component, input, output } from '@angular/core';
import { InputField } from '../input-field/input-field.component';
import { createUniqueId } from '../../utils';

@Component({
  selector: 'app-time-input-field',
  imports: [InputField],
  templateUrl: './time-input-field.component.html',
})
export class TimeInputField {
  label = input.required<string>();
  value = input.required<string>();
  id = input<string>(createUniqueId());
  placeholder = input<string>();
  disabled = input(false);
  error = input(false);
  errorText = input<string>();

  valueChange = output<string>();
}
