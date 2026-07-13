import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';

// See table-head.component.ts for why this targets a real <td> element.
@Component({
  selector: 'td[appTableCell]',
  templateUrl: './table-cell.component.html',
  host: { '[class]': 'className()' },
})
export class TableCell {
  class = input<string>();

  protected className = computed(() => classes('px-4 py-3 align-top', this.class()));
}
