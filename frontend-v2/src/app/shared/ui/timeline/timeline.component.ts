import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';

export type TimelineItem = {
  title: string;
  meta?: string;
  copy?: string;
};

@Component({
  selector: 'app-timeline',
  templateUrl: './timeline.component.html',
})
export class Timeline {
  items = input.required<TimelineItem[]>();
  class = input<string>();

  protected wrapperClass = computed(() =>
    classes('relative space-y-6 border-s border-gray-200 ps-6', this.class()),
  );
}
