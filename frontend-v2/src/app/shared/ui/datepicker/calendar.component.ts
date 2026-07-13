import { Component, computed, effect, input, output, signal } from '@angular/core';
import { classes } from '../../utils';
import {
  addMonths,
  buildMonthGrid,
  formatDisplayValue,
  formatMonthYear,
  isAfterOrEqual,
  isBeforeOrEqual,
  isSameDay,
  isSameMonth,
  parseDateValue,
  startOfMonth,
  toDateValue,
  weekdayLabels,
} from './datepicker-utils';

@Component({
  selector: 'app-calendar',
  templateUrl: './calendar.component.html',
})
export class Calendar {
  value = input<string>();
  min = input<string>();
  max = input<string>();
  class = input<string>();

  select = output<string>();

  protected readonly weekdayLabels = weekdayLabels;

  protected selectedDate = computed(() => parseDateValue(this.value()));
  private minDate = computed(() => parseDateValue(this.min()));
  private maxDate = computed(() => parseDateValue(this.max()));
  protected cursor = signal(this.selectedDate() ?? new Date());

  protected grid = computed(() => buildMonthGrid(this.cursor()));
  protected className = computed(() =>
    classes('rounded-lg border border-gray-200 bg-white p-4 shadow-xl', this.class()),
  );

  constructor() {
    effect(() => {
      const next = this.selectedDate();
      if (next) this.cursor.set(startOfMonth(next));
    });
  }

  protected monthYearLabel = computed(() => formatMonthYear(this.cursor()));
  protected valueDisplay = computed(() => formatDisplayValue(this.value(), ''));

  protected previousMonth() {
    this.cursor.update((date) => addMonths(date, -1));
  }

  protected nextMonth() {
    this.cursor.update((date) => addMonths(date, 1));
  }

  protected isDisabled(date: Date) {
    const lowerBound = this.minDate();
    const upperBound = this.maxDate();

    if (lowerBound && !isAfterOrEqual(date, lowerBound)) {
      return toDateValue(date) !== toDateValue(lowerBound);
    }

    if (upperBound && !isBeforeOrEqual(date, upperBound)) {
      return toDateValue(date) !== toDateValue(upperBound);
    }

    return false;
  }

  protected isInMonth(date: Date) {
    return isSameMonth(date, this.cursor());
  }

  protected isSelected(date: Date) {
    return isSameDay(this.selectedDate(), date);
  }

  protected isToday(date: Date) {
    return isSameDay(new Date(), date);
  }

  protected dayClass(date: Date) {
    const disabled = this.isDisabled(date);
    const inMonth = this.isInMonth(date);
    const selected = this.isSelected(date);
    const today = this.isToday(date);

    return classes(
      'rounded-md px-0 py-2 text-sm transition',
      selected
        ? 'bg-gray-950 text-white shadow-sm'
        : inMonth
          ? 'text-gray-700 hover:bg-gray-100'
          : 'text-gray-300 hover:bg-gray-50',
      today && !selected && 'ring-1 ring-gray-300',
      disabled && 'cursor-not-allowed opacity-40 hover:bg-transparent',
    );
  }

  protected selectDate(date: Date) {
    this.select.emit(toDateValue(date));
  }

  protected selectToday() {
    const today = new Date();
    this.cursor.set(startOfMonth(today));
    if (!this.isDisabled(today)) this.select.emit(toDateValue(today));
  }
}
