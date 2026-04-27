import { Card } from '../common/Primitives';

export function StatCard(props: { label: string; value: string }) {
  return (
    <Card class="space-y-2">
      <p class="text-sm font-medium uppercase tracking-wide text-gray-500">
        {props.label}
      </p>
      <p class="text-3xl font-semibold tracking-tight text-gray-900">
        {props.value}
      </p>
    </Card>
  );
}
