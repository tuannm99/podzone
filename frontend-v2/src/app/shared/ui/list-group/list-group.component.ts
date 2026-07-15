import { NgTemplateOutlet } from '@angular/common';
import { Component, computed, input } from '@angular/core';
import { MatListModule } from '@angular/material/list';
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
  imports: [RouterLink, NgTemplateOutlet, MatListModule],
  templateUrl: './list-group.component.html',
  styleUrl: './list-group.component.scss',
})
export class ListGroup {
  items = input.required<ListGroupItem[]>();
  class = input<string>();

  protected wrapperClass = computed(() => classes('app-list-group', this.class()));
}
