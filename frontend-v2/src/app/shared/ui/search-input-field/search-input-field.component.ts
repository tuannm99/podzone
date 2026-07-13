import { Component, input, output } from '@angular/core';
import { InputField } from '../input-field/input-field.component';
import { createUniqueId } from '../../utils';

@Component({
  selector: 'app-search-input-field',
  imports: [InputField],
  templateUrl: './search-input-field.component.html',
})
export class SearchInputField {
  label = input.required<string>();
  value = input.required<string>();
  id = input<string>(createUniqueId());
  placeholder = input<string>();
  disabled = input(false);
  error = input(false);
  errorText = input<string>();

  valueChange = output<string>();
}
