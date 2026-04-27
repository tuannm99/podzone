export function SectionLead(props: { eyebrow: string; title: string; copy: string }) {
  return (
    <div class="space-y-3">
      <p class="text-xs font-semibold uppercase tracking-[0.22em] text-blue-600">{props.eyebrow}</p>
      <h1 class="text-3xl font-semibold tracking-tight text-gray-900">{props.title}</h1>
      <p class="max-w-3xl text-sm text-gray-500">{props.copy}</p>
    </div>
  )
}
