import { Show } from 'solid-js'
import { classes } from '../../../shared/utils'
import { Indicator } from './Indicator'
import { avatarSizeClasses, type AvatarSize, initials } from './shared'

export function Avatar(props: {
  src?: string
  alt?: string
  name?: string
  size?: AvatarSize
  rounded?: boolean
  status?: 'online' | 'offline' | 'busy' | 'away'
  class?: string
}) {
  const statusColor = () => {
    if (props.status === 'busy') return 'red'
    if (props.status === 'away') return 'yellow'
    if (props.status === 'online') return 'green'
    return 'gray'
  }

  return (
    <div class={classes('relative inline-flex', props.class)}>
      <div
        class={classes(
          'inline-flex items-center justify-center overflow-hidden bg-gray-200 font-semibold text-gray-700',
          avatarSizeClasses[props.size ?? 'md'],
          props.rounded === false ? 'rounded-2xl' : 'rounded-full'
        )}
      >
        <Show when={props.src} fallback={<span>{initials(props.name)}</span>}>
          <img
            src={props.src}
            alt={props.alt ?? props.name ?? 'Avatar'}
            class="h-full w-full object-cover"
          />
        </Show>
      </div>
      <Show when={props.status}>
        <Indicator color={statusColor()} class="absolute bottom-0 right-0 ring-2 ring-white" />
      </Show>
    </div>
  )
}
