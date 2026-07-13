import { Component, ElementRef, effect, input, output, signal, viewChild } from '@angular/core';
import { classes } from '../../utils';
import { formatDisplayValue } from './datepicker-utils';
import { Calendar } from './calendar.component';
// NOTE: depends on Primitives.tsx's remaining exports (FieldLabel), ported
// separately (assigned to the parent agent, not this batch). Path assumed
// per this project's folder-per-component convention — reconcile at the
// consolidated build pass if the actual path differs.
import { FieldLabel } from '../field-label/field-label.component';

@Component({
  selector: 'app-datepicker-field',
  imports: [Calendar, FieldLabel],
  templateUrl: './datepicker-field.component.html',
})
export class DatepickerField {
  label = input.required<string>();
  value = input.required<string>();
  min = input<string>();
  max = input<string>();
  placeholder = input<string>();
  class = input<string>();

  valueChange = output<string>();

  protected open = signal(false);
  protected container = viewChild<ElementRef<HTMLDivElement>>('container');

  protected displayValue = () =>
    formatDisplayValue(this.value(), this.placeholder() ?? 'Select a date');
  protected wrapperClass = () => classes('relative', this.class());

  constructor() {
    effect((onCleanup) => {
      if (!this.open()) return;

      const handlePointerDown = (event: MouseEvent) => {
        const element = this.container()?.nativeElement;
        if (element && !element.contains(event.target as Node)) {
          this.open.set(false);
        }
      };

      document.addEventListener('mousedown', handlePointerDown);
      onCleanup(() => document.removeEventListener('mousedown', handlePointerDown));
    });
  }

  protected toggleOpen() {
    this.open.update((value) => !value);
  }

  protected onSelect(value: string) {
    this.valueChange.emit(value);
    this.open.set(false);
  }
}
