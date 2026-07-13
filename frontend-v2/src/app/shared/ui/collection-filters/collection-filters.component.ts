import { Component, computed, input, linkedSignal, output, signal } from '@angular/core';
import { Button } from '../button/button.component';
import type { CollectionFilter, CollectionFilterOperator } from '../../services/collection-types';

export type CollectionFilterField = {
  label: string;
  value: string;
  operators: CollectionFilterOperator[];
};

const operatorLabels: Record<CollectionFilterOperator, string> = {
  FILTER_OPERATOR_EQ: 'Equals',
  FILTER_OPERATOR_NEQ: 'Not equal',
  FILTER_OPERATOR_CONTAINS: 'Contains',
  FILTER_OPERATOR_STARTS_WITH: 'Starts with',
  FILTER_OPERATOR_GT: 'Greater than',
  FILTER_OPERATOR_GTE: 'Greater or equal',
  FILTER_OPERATOR_LT: 'Less than',
  FILTER_OPERATOR_LTE: 'Less or equal',
  FILTER_OPERATOR_IN: 'In list',
};

@Component({
  selector: 'app-collection-filters',
  imports: [Button],
  templateUrl: './collection-filters.component.html',
})
export class CollectionFilters {
  fields = input.required<CollectionFilterField[]>();
  filters = input.required<readonly CollectionFilter[]>();

  filtersChange = output<CollectionFilter[]>();

  protected readonly operatorLabels = operatorLabels;

  // linkedSignal: writable, but resets to the computation when `fields`
  // itself changes — a plain `signal()` seeded once at field-initializer
  // time can't read a required input before it's bound (NG8118).
  protected field = linkedSignal(() => this.fields()[0]?.value ?? '');
  protected selectedField = computed(() =>
    this.fields().find((item) => item.value === this.field()),
  );
  protected operator = linkedSignal<CollectionFilterOperator>(
    () => this.selectedField()?.operators[0] ?? 'FILTER_OPERATOR_EQ',
  );
  protected value = signal('');

  protected selectField(nextField: string) {
    this.field.set(nextField);
    const next = this.fields().find((item) => item.value === nextField);
    this.operator.set(next?.operators[0] ?? 'FILTER_OPERATOR_EQ');
  }

  protected onFieldChange(event: Event) {
    this.selectField((event.target as HTMLSelectElement).value);
  }

  protected onOperatorChange(event: Event) {
    this.operator.set((event.target as HTMLSelectElement).value as CollectionFilterOperator);
  }

  protected onValueInput(event: Event) {
    this.value.set((event.target as HTMLInputElement).value);
  }

  protected addFilter(event: SubmitEvent) {
    event.preventDefault();
    const normalized = this.value().trim();
    const field = this.field();
    if (!field || !normalized) return;
    const operator = this.operator();
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
}
