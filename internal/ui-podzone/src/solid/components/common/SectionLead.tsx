export function SectionLead(props: {
  eyebrow: string
  title: string
  copy: string
}) {
  return (
    <div class="space-y-2">
      <p class="text-xs font-semibold uppercase text-gray-500">
        {props.eyebrow}
      </p>
      <h1 class="text-2xl font-semibold text-gray-950">{props.title}</h1>
      <p class="max-w-3xl text-sm leading-6 text-gray-600">{props.copy}</p>
    </div>
  )
}
