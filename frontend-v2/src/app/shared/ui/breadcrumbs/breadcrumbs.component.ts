import { Component, input } from '@angular/core';
import { classes } from '../../utils';
import { Link } from '../link/link.component';

export type BreadcrumbItem = {
  label: string;
  href?: string;
  current?: boolean;
};

@Component({
  selector: 'app-breadcrumbs',
  imports: [Link],
  templateUrl: './breadcrumbs.component.html',
})
export class Breadcrumbs {
  items = input.required<BreadcrumbItem[]>();
  class = input<string>();

  protected isCurrent(item: BreadcrumbItem, index: number) {
    return item.current || index === this.items().length - 1;
  }

  protected labelClass(item: BreadcrumbItem, index: number) {
    return classes('font-medium', this.isCurrent(item, index) ? 'text-gray-900' : 'text-gray-600');
  }
}
