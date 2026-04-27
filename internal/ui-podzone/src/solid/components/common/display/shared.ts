export type SurfaceTone = 'blue' | 'green' | 'yellow' | 'red' | 'dark'
export type AlertTone = SurfaceTone
export type AvatarSize = 'xs' | 'sm' | 'md' | 'lg' | 'xl'
export type StepStatus = 'complete' | 'current' | 'upcoming'

export const toneClasses: Record<SurfaceTone, string> = {
  blue: 'border-blue-200 bg-blue-50 text-blue-900',
  green: 'border-green-200 bg-green-50 text-green-900',
  yellow: 'border-amber-200 bg-amber-50 text-amber-900',
  red: 'border-red-200 bg-red-50 text-red-900',
  dark: 'border-gray-200 bg-gray-100 text-gray-900'
}

export const toneFillClasses: Record<SurfaceTone, string> = {
  blue: 'bg-blue-600',
  green: 'bg-green-600',
  yellow: 'bg-amber-500',
  red: 'bg-red-600',
  dark: 'bg-gray-700'
}

export const indicatorClasses: Record<'blue' | 'green' | 'yellow' | 'red' | 'gray', string> = {
  blue: 'bg-blue-500',
  green: 'bg-green-500',
  yellow: 'bg-amber-400',
  red: 'bg-red-500',
  gray: 'bg-gray-400'
}

export const avatarSizeClasses: Record<AvatarSize, string> = {
  xs: 'size-8 text-xs',
  sm: 'size-10 text-sm',
  md: 'size-12 text-base',
  lg: 'size-16 text-lg',
  xl: 'size-20 text-xl'
}

export function initials(name: string | undefined) {
  if (!name) return '?'
  return name
    .trim()
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? '')
    .join('')
}
