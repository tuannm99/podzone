import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';
import { type StepStatus } from '../display-shared';

export type StepperItem = {
  title: string;
  description?: string;
  status?: StepStatus;
};

@Component({
  selector: 'app-stepper',
  templateUrl: './stepper.component.html',
})
export class Stepper {
  items = input.required<StepperItem[]>();
  class = input<string>();

  protected wrapperClass = computed(() => classes('space-y-4', this.class()));

  protected dotClass(item: StepperItem) {
    const status = item.status ?? 'upcoming';
    return classes(
      'flex size-8 items-center justify-center rounded-full text-xs font-semibold',
      status === 'complete'
        ? 'bg-green-600 text-white'
        : status === 'current'
          ? 'bg-blue-600 text-white'
          : 'bg-gray-200 text-gray-600',
    );
  }
}
