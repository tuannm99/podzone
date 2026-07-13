import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';

// See table-head.component.ts for why this targets a real <th> element.
@Component({
  selector: 'th[appTableHeaderCell]',
  templateUrl: './table-header-cell.component.html',
  host: { '[class]': 'className()' },
})
export class TableHeaderCell {
  class = input<string>();

  protected className = computed(() => classes('px-4 py-3 font-medium', this.class()));
}
