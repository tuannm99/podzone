import { Component, input } from '@angular/core';

// See table-head.component.ts for why this targets a real <tr> element.
@Component({
  selector: 'tr[appTableRow]',
  templateUrl: './table-row.component.html',
  host: { '[class]': 'class()' },
})
export class TableRow {
  class = input<string>();
}
