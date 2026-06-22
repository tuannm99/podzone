import { Card } from '../common/Primitives'

export function StatCard(props: { label: string; value: string }) {
  return (
    <Card class="space-y-1">
      <p class="text-xs font-semibold uppercase text-gray-500">{props.label}</p>
      <p class="text-2xl font-semibold text-gray-950">{props.value}</p>
    </Card>
  )
}
