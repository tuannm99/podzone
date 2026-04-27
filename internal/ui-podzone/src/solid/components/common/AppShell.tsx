import type { ParentProps } from 'solid-js'
import { classes } from '../../shared/utils'

type ContainerWidth = 'lg' | 'xl' | '2xl' | '7xl'

const widthClasses: Record<ContainerWidth, string> = {
  lg: 'max-w-5xl',
  xl: 'max-w-6xl',
  '2xl': 'max-w-[96rem]',
  '7xl': 'max-w-7xl'
}

export function Container(props: ParentProps<{ class?: string; width?: ContainerWidth }>) {
  return (
    <div
      class={classes(
        'mx-auto w-full px-4 sm:px-6 lg:px-8',
        widthClasses[props.width ?? '7xl'],
        props.class
      )}
    >
      {props.children}
    </div>
  )
}

export function AppShell(props: ParentProps<{ class?: string; containerClass?: string }>) {
  return (
    <div class={classes('min-h-screen bg-gray-50 text-gray-900', props.class)}>
      <Container class={classes('pb-6 pt-0', props.containerClass)}>{props.children}</Container>
    </div>
  )
}
