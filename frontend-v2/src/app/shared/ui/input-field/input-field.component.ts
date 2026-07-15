import { Component, computed, input, output } from '@angular/core';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { createUniqueId } from '../../utils';

@Component({
  selector: 'app-input-field',
  imports: [MatFormFieldModule, MatInputModule],
  templateUrl: './input-field.component.html',
  styleUrl: './input-field.component.scss',
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

  protected errorId = computed(() => (this.errorText() ? `${this.id()}-error` : null));
}
