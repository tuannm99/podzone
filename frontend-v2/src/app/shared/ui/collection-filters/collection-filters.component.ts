import { Component, computed, input, output, signal } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatChipsModule } from '@angular/material/chips';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';

import {
  operatorLabels,
  operatorsFor,
  type CollectionFieldSpec,
} from '../../services/collection-field-spec';
import type { CollectionFilter, CollectionFilterOperator } from '../../services/collection-types';

// Port of frontend/packages/shared/ui/components/common/CollectionFilters.tsx,
// adapted per ADR-0008: operators come from `operatorsFor(field.kind)`
// instead of a hand-written `operators` array per field, so a filterable
// field can never offer an operator its kind doesn't support (e.g.
// `contains` on a timestamp field) - the same guard the backend's
// pkg/collection/sqlspec.Spec enforces, applied here before the request
// ever leaves the browser.
@Component({
  selector: 'app-collection-filters',
  imports: [
    MatFormFieldModule,
    MatSelectModule,
    MatInputModule,
    MatButtonModule,
    MatIconModule,
    MatChipsModule,
  ],
  templateUrl: './collection-filters.component.html',
  styleUrl: './collection-filters.component.scss',
})
export class CollectionFilters<T> {
  fields = input.required<CollectionFieldSpec<T>[]>();
  filters = input<CollectionFilter[]>([]);

  filtersChange = output<CollectionFilter[]>();

  protected filterableFields = computed(() => this.fields().filter((field) => field.filterable));

  private selectedFieldOverride = signal<string | null>(null);
  protected selectedOperator = signal<CollectionFilterOperator | null>(null);
  protected value = signal('');

  // Defaults to the first filterable field until the user picks one -
  // avoids seeding from `fields()` at construction time, which isn't
  // guaranteed to reflect the real bound input yet.
  protected selectedField = computed(
    () => this.selectedFieldOverride() ?? this.filterableFields()[0]?.field ?? '',
  );

  protected operatorOptions = computed(() => {
    const field = this.filterableFields().find((item) => item.field === this.selectedField());
    return field ? operatorsFor(field.kind) : [];
  });

  protected effectiveOperator = computed(
    () => this.selectedOperator() ?? this.operatorOptions()[0] ?? 'FILTER_OPERATOR_EQ',
  );

  protected readonly operatorLabels = operatorLabels;

  protected onFieldChange(field: string) {
    this.selectedFieldOverride.set(field);
    this.selectedOperator.set(null);
  }

  protected onSubmit(event: SubmitEvent) {
    event.preventDefault();
    const field = this.selectedField();
    const normalized = this.value().trim();
    if (!field || !normalized) return;

    const operator = this.effectiveOperator();
    const values =
      operator === 'FILTER_OPERATOR_IN'
        ? normalized
            .split(',')
            .map((item) => item.trim())
            .filter(Boolean)
        : [normalized];

    this.filtersChange.emit([
      ...this.filters().filter((item) => item.field !== field),
      { field, operator, values },
    ]);
    this.value.set('');
  }

  protected removeFilter(index: number) {
    this.filtersChange.emit(this.filters().filter((_, itemIndex) => itemIndex !== index));
  }

  protected labelFor(field: string): string {
    return this.fields().find((item) => item.field === field)?.label ?? field;
  }
}
