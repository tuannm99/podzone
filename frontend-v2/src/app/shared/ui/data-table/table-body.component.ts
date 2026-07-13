import { Component } from '@angular/core';

// See table-head.component.ts for why this targets a real <tbody> element
// instead of wrapping it in a custom tag.
@Component({
  selector: 'tbody[appTableBody]',
  templateUrl: './table-body.component.html',
  host: { class: 'divide-y divide-gray-200 bg-white' },
})
export class TableBody {}
