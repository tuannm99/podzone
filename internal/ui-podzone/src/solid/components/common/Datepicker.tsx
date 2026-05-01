import {
  For,
  Show,
  createEffect,
  createMemo,
  createSignal,
  onCleanup,
} from 'solid-js';
import { classes } from '../../shared/utils';
import { FieldLabel } from './Primitives';

const weekdayLabels = ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa'];

function pad(value: number) {
  return String(value).padStart(2, '0');
}

function toDateValue(date: Date) {
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`;
}

function parseDateValue(value: string | undefined) {
  if (!value || !/^\d{4}-\d{2}-\d{2}$/.test(value)) return null;
  const [year, month, day] = value.split('-').map(Number);
  const date = new Date(year, month - 1, day);
  if (
    date.getFullYear() !== year ||
    date.getMonth() !== month - 1 ||
    date.getDate() !== day
  ) {
    return null;
  }
  return date;
}

function addMonths(date: Date, amount: number) {
  return new Date(date.getFullYear(), date.getMonth() + amount, 1);
}

function startOfMonth(date: Date) {
  return new Date(date.getFullYear(), date.getMonth(), 1);
}

function isSameDay(left: Date | null, right: Date) {
  return (
    !!left &&
    left.getFullYear() === right.getFullYear() &&
    left.getMonth() === right.getMonth() &&
    left.getDate() === right.getDate()
  );
}

function isSameMonth(left: Date, right: Date) {
  return (
    left.getFullYear() === right.getFullYear() &&
    left.getMonth() === right.getMonth()
  );
}

function buildMonthGrid(cursor: Date) {
  const firstDay = startOfMonth(cursor);
  const gridStart = new Date(firstDay);
  gridStart.setDate(firstDay.getDate() - firstDay.getDay());

  return Array.from({ length: 42 }, (_, index) => {
    const date = new Date(gridStart);
    date.setDate(gridStart.getDate() + index);
    return date;
  });
}

function formatMonthYear(date: Date) {
  return date.toLocaleDateString(undefined, {
    month: 'long',
    year: 'numeric',
  });
}

function formatDisplayValue(value: string | undefined, placeholder: string) {
  const parsed = parseDateValue(value);
  if (!parsed) return placeholder;
  return parsed.toLocaleDateString();
}

function isBeforeOrEqual(left: Date, right: Date) {
  return toDateValue(left) <= toDateValue(right);
}

function isAfterOrEqual(left: Date, right: Date) {
  return toDateValue(left) >= toDateValue(right);
}

export function Calendar(props: {
  value?: string;
  min?: string;
  max?: string;
  class?: string;
  onSelect: (value: string) => void;
}) {
  const selectedDate = createMemo(() => parseDateValue(props.value));
  const minDate = createMemo(() => parseDateValue(props.min));
  const maxDate = createMemo(() => parseDateValue(props.max));
  const [cursor, setCursor] = createSignal(selectedDate() ?? new Date());

  createEffect(() => {
    const next = selectedDate();
    if (next) setCursor(startOfMonth(next));
  });

  const grid = createMemo(() => buildMonthGrid(cursor()));

  const isDisabled = (date: Date) => {
    const lowerBound = minDate();
    const upperBound = maxDate();

    if (lowerBound && !isAfterOrEqual(date, lowerBound)) {
      return toDateValue(date) !== toDateValue(lowerBound);
    }

    if (upperBound && !isBeforeOrEqual(date, upperBound)) {
      return toDateValue(date) !== toDateValue(upperBound);
    }

    return false;
  };

  return (
    <div
      class={classes(
        'rounded-2xl border border-gray-200 bg-white p-4 shadow-xl',
        props.class
      )}
    >
      <div class="flex items-center justify-between gap-3">
        <button
          type="button"
          class="rounded-xl px-3 py-2 text-sm font-medium text-gray-600 transition hover:bg-gray-100 hover:text-gray-900"
          onClick={() => setCursor((date) => addMonths(date, -1))}
          aria-label="Previous month"
        >
          ←
        </button>
        <p class="text-sm font-semibold text-gray-900">
          {formatMonthYear(cursor())}
        </p>
        <button
          type="button"
          class="rounded-xl px-3 py-2 text-sm font-medium text-gray-600 transition hover:bg-gray-100 hover:text-gray-900"
          onClick={() => setCursor((date) => addMonths(date, 1))}
          aria-label="Next month"
        >
          →
        </button>
      </div>

      <div class="mt-4 grid grid-cols-7 gap-2">
        <For each={weekdayLabels}>
          {(label) => (
            <span class="text-center text-xs font-semibold uppercase tracking-wide text-gray-400">
              {label}
            </span>
          )}
        </For>

        <For each={grid()}>
          {(date) => {
            const disabled = () => isDisabled(date);
            const inMonth = () => isSameMonth(date, cursor());
            const selected = () => isSameDay(selectedDate(), date);
            const today = () => isSameDay(new Date(), date);

            return (
              <button
                type="button"
                disabled={disabled()}
                class={classes(
                  'rounded-xl px-0 py-2 text-sm transition',
                  selected()
                    ? 'bg-blue-700 text-white shadow-sm'
                    : inMonth()
                      ? 'text-gray-700 hover:bg-gray-100'
                      : 'text-gray-300 hover:bg-gray-50',
                  today() && !selected() && 'ring-1 ring-blue-200',
                  disabled() &&
                    'cursor-not-allowed opacity-40 hover:bg-transparent'
                )}
                onClick={() => props.onSelect(toDateValue(date))}
              >
                {date.getDate()}
              </button>
            );
          }}
        </For>
      </div>

      <div class="mt-4 flex items-center justify-between gap-3 border-t border-gray-100 pt-4">
        <button
          type="button"
          class="rounded-xl px-3 py-2 text-sm font-medium text-gray-600 transition hover:bg-gray-100 hover:text-gray-900"
          onClick={() => {
            const today = new Date();
            setCursor(startOfMonth(today));
            if (!isDisabled(today)) props.onSelect(toDateValue(today));
          }}
        >
          Today
        </button>
        <Show when={props.value}>
          <span class="text-sm text-gray-500">
            {formatDisplayValue(props.value, '')}
          </span>
        </Show>
      </div>
    </div>
  );
}

export function DatepickerField(props: {
  label: string;
  value: string;
  min?: string;
  max?: string;
  placeholder?: string;
  class?: string;
  onChange: (value: string) => void;
}) {
  const [open, setOpen] = createSignal(false);
  let container: HTMLDivElement | undefined;

  createEffect(() => {
    if (!open()) return;

    const handlePointerDown = (event: MouseEvent) => {
      if (container && !container.contains(event.target as Node)) {
        setOpen(false);
      }
    };

    document.addEventListener('mousedown', handlePointerDown);
    onCleanup(() => {
      document.removeEventListener('mousedown', handlePointerDown);
    });
  });

  return (
    <FieldLabel label={props.label}>
      <div class={classes('relative', props.class)} ref={container}>
        <button
          type="button"
          class="flex w-full items-center justify-between rounded-xl border border-gray-300 bg-white px-3 py-2.5 text-left text-sm text-gray-900 shadow-sm outline-none transition hover:border-gray-400 focus:border-blue-500 focus:ring-4 focus:ring-blue-100"
          onClick={() => setOpen((value) => !value)}
        >
          <span class={props.value ? 'text-gray-900' : 'text-gray-400'}>
            {formatDisplayValue(
              props.value,
              props.placeholder ?? 'Select a date'
            )}
          </span>
          <span class="text-gray-400">⌄</span>
        </button>

        <Show when={open()}>
          <div class="absolute left-0 top-full z-30 mt-2 w-full min-w-[19rem]">
            <Calendar
              value={props.value}
              min={props.min}
              max={props.max}
              onSelect={(value) => {
                props.onChange(value);
                setOpen(false);
              }}
            />
          </div>
        </Show>
      </div>
    </FieldLabel>
  );
}

export const Datepicker = DatepickerField;
