import { NgTemplateOutlet } from '@angular/common';
import { Component, computed, input } from '@angular/core';
import { RouterLink } from '@angular/router';
import { classes } from '../../utils';

// NOTE (port judgment call): Solid's ListGroupItem.prefix/suffix are
// `JSX.Element` — arbitrary icon/content slots per array item. Angular has
// no direct equivalent without a TemplateRef-per-item API (a bigger
// structural change than a faithful port). Dropped here; label/description/
// href/active/onClick are ported. Add a TemplateRef-based prefix/suffix API
// later if a consumer actually needs per-item icons.
export type ListGroupItem = {
  label: string;
  description?: string;
  href?: string;
  active?: boolean;
  onClick?: () => void;
};

@Component({
  selector: 'app-list-group',
  imports: [RouterLink, NgTemplateOutlet],
  templateUrl: './list-group.component.html',
})
export class ListGroup {
  items = input.required<ListGroupItem[]>();
  class = input<string>();

  protected wrapperClass = computed(() =>
    classes('overflow-hidden rounded-lg border border-gray-200 bg-white shadow-sm', this.class()),
  );

  protected itemClass(item: ListGroupItem) {
    return classes(
      'flex w-full items-center justify-between gap-4 border-b border-gray-100 px-4 py-3 text-left transition last:border-b-0',
      item.active ? 'bg-blue-50 text-blue-900' : 'text-gray-700 hover:bg-gray-50',
    );
  }
}
