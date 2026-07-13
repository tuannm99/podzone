import { Component, TemplateRef, computed, input } from '@angular/core';
import { NgTemplateOutlet } from '@angular/common';
import { classes } from '../../utils';

@Component({
  selector: 'app-prose-list',
  imports: [NgTemplateOutlet],
  templateUrl: './prose-list.component.html',
})
export class ProseList {
  items = input.required<Array<string | TemplateRef<unknown>>>();
  ordered = input(false);
  class = input<string>();

  protected className = computed(() =>
    classes(
      'space-y-2 ps-5 text-base leading-7 text-gray-700',
      this.ordered() ? 'list-decimal' : 'list-disc',
      this.class(),
    ),
  );

  protected isString(item: string | TemplateRef<unknown>): item is string {
    return typeof item === 'string';
  }
}
