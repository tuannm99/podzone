import { Component } from '@angular/core';

// Selector targets a real <thead> element (not a wrapping custom element)
// because HTML table parsing foster-parents any non-table-content element
// (like a custom <app-table-head> tag) out of the <table> entirely — a
// wrapper-component wouldn't render inside the table at all. Usage:
// <thead appTableHead>...</thead>, matching Solid's <TableHead> which
// rendered a real <thead> directly.
@Component({
  selector: 'thead[appTableHead]',
  templateUrl: './table-head.component.html',
  host: { class: 'bg-gray-50 text-gray-600' },
})
export class TableHead {}
