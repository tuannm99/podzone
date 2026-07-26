import { Component, input, output, signal } from '@angular/core';
import { MatIconModule } from '@angular/material/icon';

// Port of frontend/packages/shared/ui/components/common/CollectionToolbar.tsx,
// trimmed to search only: sort now lives on app-data-table's clickable
// column headers (mat-sort-header), not a separate dropdown here, and page
// size is app-data-table's mat-paginator's job — so this component owns
// exactly one concern.
//
// Uses a plain <input>, not mat-form-field: mat-form-field's outlined
// appearance computes its notched-label layout from an internal height it
// expects to own. A previous attempt to shrink it via a forced
// ::ng-deep height override broke that notch calculation (the label text
// rendered detached from/overlapping the border instead of sitting inside
// it). A plain input styled directly gives full, predictable control over
// a compact size without fighting Material internals we can't see rendered
// without a browser.
@Component({
  selector: 'app-collection-toolbar',
  imports: [MatIconModule],
  templateUrl: './collection-toolbar.component.html',
  styleUrl: './collection-toolbar.component.scss',
})
export class CollectionToolbar {
  search = input('');
  searchPlaceholder = input('Search');

  searchChange = output<string>();

  protected searchDraft = signal(this.search());

  protected onSubmit(event: SubmitEvent) {
    event.preventDefault();
    this.searchChange.emit(this.searchDraft().trim());
  }
}
