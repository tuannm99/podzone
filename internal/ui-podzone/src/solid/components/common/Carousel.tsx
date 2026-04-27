import { For, Show, createEffect, createSignal, onCleanup, type JSX } from 'solid-js'
import { classes } from '../../shared/utils'

export type CarouselSlide = {
  id?: string
  eyebrow?: string
  title?: string
  copy?: string
  imageSrc?: string
  imageAlt?: string
  content?: JSX.Element
  action?: JSX.Element
}

export function Carousel(props: {
  slides: CarouselSlide[]
  autoPlay?: boolean
  intervalMs?: number
  class?: string
}) {
  const [currentIndex, setCurrentIndex] = createSignal(0)

  createEffect(() => {
    if (!props.autoPlay || props.slides.length <= 1) return

    const timer = window.setInterval(() => {
      setCurrentIndex((index) => (index + 1) % props.slides.length)
    }, props.intervalMs ?? 5000)

    onCleanup(() => {
      window.clearInterval(timer)
    })
  })

  const goTo = (index: number) => {
    setCurrentIndex(index)
  }

  const previous = () => {
    setCurrentIndex((index) => (index - 1 + props.slides.length) % props.slides.length)
  }

  const next = () => {
    setCurrentIndex((index) => (index + 1) % props.slides.length)
  }

  return (
    <section
      class={classes(
        'overflow-hidden rounded-3xl border border-gray-200 bg-white shadow-sm',
        props.class
      )}
      aria-roledescription="carousel"
    >
      <Show
        when={props.slides.length > 0}
        fallback={<div class="px-6 py-10 text-sm text-gray-500">No slides configured.</div>}
      >
        <div class="relative min-h-[24rem]">
          <For each={props.slides}>
            {(slide, index) => (
              <Show when={index() === currentIndex()}>
                <article class="grid min-h-[24rem] gap-8 p-6 lg:grid-cols-[minmax(0,1.1fr)_minmax(20rem,0.9fr)] lg:p-8">
                  <div class="flex flex-col justify-center space-y-4">
                    <Show when={slide.eyebrow}>
                      <p class="text-xs font-semibold uppercase tracking-[0.24em] text-blue-700">
                        {slide.eyebrow}
                      </p>
                    </Show>
                    <Show when={slide.title}>
                      <h2 class="text-3xl font-semibold tracking-tight text-gray-900 sm:text-4xl">
                        {slide.title}
                      </h2>
                    </Show>
                    <Show when={slide.copy}>
                      <p class="max-w-2xl text-base leading-7 text-gray-600">{slide.copy}</p>
                    </Show>
                    <Show when={slide.content}>
                      <div>{slide.content}</div>
                    </Show>
                    <Show when={slide.action}>
                      <div class="pt-2">{slide.action}</div>
                    </Show>
                  </div>

                  <Show when={slide.imageSrc}>
                    <div class="relative overflow-hidden rounded-3xl bg-gradient-to-br from-blue-50 to-white">
                      <img
                        src={slide.imageSrc}
                        alt={slide.imageAlt ?? slide.title ?? 'Carousel image'}
                        class="h-full w-full object-cover"
                      />
                    </div>
                  </Show>
                </article>
              </Show>
            )}
          </For>

          <Show when={props.slides.length > 1}>
            <div class="pointer-events-none absolute inset-x-0 bottom-0 flex items-center justify-between gap-4 px-6 pb-6">
              <div class="pointer-events-auto flex items-center gap-2 rounded-full bg-white/90 px-3 py-2 shadow-lg ring-1 ring-gray-200 backdrop-blur">
                <button
                  type="button"
                  class="rounded-full px-3 py-2 text-sm font-medium text-gray-700 transition hover:bg-gray-100"
                  onClick={previous}
                  aria-label="Previous slide"
                >
                  ←
                </button>
                <button
                  type="button"
                  class="rounded-full px-3 py-2 text-sm font-medium text-gray-700 transition hover:bg-gray-100"
                  onClick={next}
                  aria-label="Next slide"
                >
                  →
                </button>
              </div>

              <div class="pointer-events-auto flex items-center gap-2 rounded-full bg-white/90 px-3 py-2 shadow-lg ring-1 ring-gray-200 backdrop-blur">
                <For each={props.slides}>
                  {(_, index) => (
                    <button
                      type="button"
                      class={classes(
                        'size-2.5 rounded-full transition',
                        index() === currentIndex() ? 'bg-blue-700' : 'bg-gray-300 hover:bg-gray-400'
                      )}
                      onClick={() => goTo(index())}
                      aria-label={`Go to slide ${index() + 1}`}
                    />
                  )}
                </For>
              </div>
            </div>
          </Show>
        </div>
      </Show>
    </section>
  )
}
