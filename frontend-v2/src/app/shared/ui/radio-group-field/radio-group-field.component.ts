import { Component, input, output } from '@angular/core';

export type RadioOption = {
  label: string;
  value: string;
  hint?: string;
};

@Component({
  selector: 'app-radio-group-field',
  templateUrl: './radio-group-field.component.html',
})
export class RadioGroupField {
  label = input.required<string>();
  name = input.required<string>();
  value = input.required<string>();
  options = input.required<RadioOption[]>();

  valueChange = output<string>();
}
