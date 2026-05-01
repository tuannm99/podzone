import type { ParentProps } from 'solid-js';
import { classes } from '../../shared/utils';

export function DataTable(props: ParentProps<{ class?: string }>) {
  return (
    <div
      class={classes(
        'overflow-x-auto rounded-2xl border border-gray-200',
        props.class
      )}
    >
      <table class="min-w-full divide-y divide-gray-200 text-left text-sm">
        {props.children}
      </table>
    </div>
  );
}

export function TableHead(props: ParentProps) {
  return <thead class="bg-gray-50 text-gray-600">{props.children}</thead>;
}

export function TableBody(props: ParentProps) {
  return (
    <tbody class="divide-y divide-gray-200 bg-white">{props.children}</tbody>
  );
}

export function TableRow(props: ParentProps) {
  return <tr>{props.children}</tr>;
}

export function TableHeaderCell(props: ParentProps<{ class?: string }>) {
  return (
    <th class={classes('px-4 py-3 font-medium', props.class)}>
      {props.children}
    </th>
  );
}

export function TableCell(props: ParentProps<{ class?: string }>) {
  return (
    <td class={classes('px-4 py-3 align-top', props.class)}>
      {props.children}
    </td>
  );
}
