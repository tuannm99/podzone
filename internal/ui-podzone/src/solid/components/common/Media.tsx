import { For } from 'solid-js'
import { classes } from '../../shared/utils'

type GalleryItem = {
  src: string
  alt: string
  caption?: string
}

const columnClasses = {
  2: 'sm:grid-cols-2',
  3: 'sm:grid-cols-2 lg:grid-cols-3',
  4: 'sm:grid-cols-2 lg:grid-cols-4'
} as const

export function GalleryGrid(props: { items: GalleryItem[]; columns?: 2 | 3 | 4; class?: string }) {
  return (
    <div class={classes('grid gap-4', columnClasses[props.columns ?? 3], props.class)}>
      <For each={props.items}>
        {(item) => (
          <figure class="overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-sm">
            <img src={item.src} alt={item.alt} class="h-56 w-full object-cover" />
            {item.caption ? (
              <figcaption class="px-4 py-3 text-sm text-gray-500">{item.caption}</figcaption>
            ) : null}
          </figure>
        )}
      </For>
    </div>
  )
}

export function VideoEmbed(props: {
  title: string
  src: string
  aspect?: 'video' | 'wide' | 'square'
  class?: string
}) {
  const aspectClass = () => {
    if (props.aspect === 'square') return 'aspect-square'
    if (props.aspect === 'wide') return 'aspect-[21/9]'
    return 'aspect-video'
  }

  return (
    <div
      class={classes(
        'overflow-hidden rounded-2xl border border-gray-200 bg-black shadow-sm',
        props.class
      )}
    >
      <iframe
        src={props.src}
        title={props.title}
        allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
        allowfullscreen
        class={classes('w-full', aspectClass())}
      />
    </div>
  )
}
