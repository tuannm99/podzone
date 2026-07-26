import { Component, input, output } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatIconModule } from '@angular/material/icon';
import { MatMenuModule } from '@angular/material/menu';

export type ColumnPickerItem = { key: string; header: string };

// Extracted from app-data-table so it can sit in the same controls row as
// app-collection-toolbar/app-collection-filters (previously it rendered
// inside DataTable, isolated in its own row directly above the table with
// no visual relationship to the rest of the controls). DataTable stays the
// source of truth for which columns *exist*; this component only tracks
// which are hidden, and the page wires the two together via
// app-data-table's `hiddenColumns` input.
@Component({
  selector: 'app-column-picker',
  imports: [MatButtonModule, MatIconModule, MatMenuModule, MatCheckboxModule],
  templateUrl: './column-picker.component.html',
  styleUrl: './column-picker.component.scss',
})
export class ColumnPicker {
  columns = input.required<ColumnPickerItem[]>();
  hidden = input<ReadonlySet<string>>(new Set());

  toggle = output<string>();

  protected isVisible(key: string): boolean {
    return !this.hidden().has(key);
  }
}
