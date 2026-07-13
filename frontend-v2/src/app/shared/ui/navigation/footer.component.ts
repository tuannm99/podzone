import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';
import { NavAction } from './nav-action.component';
import type { NavItem } from './nav-item';

@Component({
  selector: 'app-footer',
  imports: [NavAction],
  templateUrl: './footer.component.html',
})
export class Footer {
  links = input<NavItem[]>([]);
  note = input<string>();
  class = input<string>();
  // Solid's `brand` (JSX.Element) becomes a content-projection slot —
  // use <div slot="brand"> inside <app-footer>. Default children
  // (Solid's `props.children`) map to the unslotted <ng-content>.

  protected className = computed(() =>
    classes('rounded-lg border border-gray-200 bg-white px-6 py-5 shadow-sm', this.class()),
  );
}
