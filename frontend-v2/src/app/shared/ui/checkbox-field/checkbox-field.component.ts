import { Component, input, output } from '@angular/core';

@Component({
  selector: 'app-checkbox-field',
  templateUrl: './checkbox-field.component.html',
})
export class CheckboxField {
  label = input.required<string>();
  checked = input.required<boolean>();

  checkedChange = output<boolean>();
}
