import { Component, input, output } from '@angular/core';

@Component({
  selector: 'app-toggle-field',
  templateUrl: './toggle-field.component.html',
})
export class ToggleField {
  label = input.required<string>();
  checked = input.required<boolean>();

  checkedChange = output<boolean>();
}
