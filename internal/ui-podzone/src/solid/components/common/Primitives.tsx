import { For, Show, splitProps, type JSX, type ParentProps } from 'solid-js'
import { classes } from '../../shared/utils'

type ButtonColor = 'blue' | 'alternative' | 'light' | 'dark' | 'green' | 'red'
type ButtonSize = 'xs' | 'sm' | 'md'
type BadgeColor = 'blue' | 'indigo' | 'green' | 'yellow' | 'pink' | 'dark' | 'red'

const buttonColorClasses: Record<ButtonColor, string> = {
  blue: 'bg-blue-700 text-white hover:bg-blue-800 focus:ring-blue-300',
  alternative:
    'border border-gray-200 bg-white text-gray-900 hover:bg-gray-100 focus:ring-gray-200',
  light: 'border border-gray-200 bg-white text-gray-700 hover:bg-gray-50 focus:ring-gray-200',
  dark: 'bg-gray-800 text-white hover:bg-gray-900 focus:ring-gray-300',
  green: 'bg-green-700 text-white hover:bg-green-800 focus:ring-green-300',
  red: 'bg-red-700 text-white hover:bg-red-800 focus:ring-red-300'
}

const buttonSizeClasses: Record<ButtonSize, string> = {
  xs: 'px-3 py-2 text-xs',
  sm: 'px-4 py-2 text-sm',
  md: 'px-5 py-2.5 text-sm'
}

const badgeColorClasses: Record<BadgeColor, string> = {
  blue: 'bg-blue-100 text-blue-800',
  indigo: 'bg-indigo-100 text-indigo-800',
  green: 'bg-green-100 text-green-800',
  yellow: 'bg-yellow-100 text-yellow-800',
  pink: 'bg-pink-100 text-pink-800',
  dark: 'bg-gray-100 text-gray-800',
  red: 'bg-red-100 text-red-800'
}

function fieldBaseClasses(hasError?: boolean) {
  return classes(
    'block w-full rounded-xl border bg-white px-3 py-2.5 text-sm text-gray-900 shadow-sm outline-none transition',
    hasError
      ? 'border-red-300 focus:border-red-500 focus:ring-4 focus:ring-red-100'
      : 'border-gray-300 focus:border-blue-500 focus:ring-4 focus:ring-blue-100'
  )
}

export function Spinner(props: { class?: string }) {
  return (
    <span
      class={classes(
        'inline-block size-4 animate-spin rounded-full border-2 border-current border-r-transparent',
        props.class
      )}
      aria-hidden="true"
    />
  )
}

export function Card(props: ParentProps<{ class?: string }>) {
  return (
    <section
      class={classes('rounded-2xl border border-gray-200 bg-white p-6 shadow-sm', props.class)}
    >
      {props.children}
    </section>
  )
}

export function Badge(props: { content: string; color?: BadgeColor; class?: string }) {
  return (
    <span
      class={classes(
        'inline-flex items-center rounded-full px-2.5 py-1 text-xs font-semibold',
        badgeColorClasses[props.color ?? 'dark'],
        props.class
      )}
    >
      {props.content}
    </span>
  )
}

type ButtonProps = ParentProps<{
  color?: ButtonColor
  size?: ButtonSize
  pill?: boolean
  href?: string
  target?: string
  type?: 'button' | 'submit' | 'reset'
  loading?: boolean
  disabled?: boolean
  class?: string
  onClick?: JSX.EventHandlerUnion<HTMLButtonElement | HTMLAnchorElement, MouseEvent>
}>

export function Button(props: ButtonProps) {
  const [local, rest] = splitProps(props, [
    'children',
    'color',
    'size',
    'pill',
    'href',
    'target',
    'type',
    'loading',
    'disabled',
    'class',
    'onClick'
  ])

  const className = classes(
    'inline-flex items-center justify-center gap-2 font-medium focus:outline-none focus:ring-4 disabled:pointer-events-none disabled:opacity-60',
    buttonColorClasses[local.color ?? 'blue'],
    buttonSizeClasses[local.size ?? 'md'],
    local.pill ? 'rounded-full' : 'rounded-xl',
    local.class
  )

  const content = (
    <>
      <Show when={local.loading}>
        <Spinner class="size-3.5" />
      </Show>
      {local.children}
    </>
  )

  return local.href ? (
    <a
      href={local.href}
      target={local.target}
      class={className}
      aria-disabled={local.disabled || local.loading}
      onClick={
        local.disabled || local.loading
          ? (event) => event.preventDefault()
          : (local.onClick as JSX.EventHandlerUnion<HTMLAnchorElement, MouseEvent>)
      }
      {...rest}
    >
      {content}
    </a>
  ) : (
    <button
      type={local.type ?? 'button'}
      class={className}
      disabled={local.disabled || local.loading}
      onClick={local.onClick as JSX.EventHandlerUnion<HTMLButtonElement, MouseEvent>}
      {...rest}
    >
      {content}
    </button>
  )
}

export function FieldLabel(props: ParentProps<{ label: string }>) {
  return (
    <label class="space-y-2">
      <span class="block text-sm font-medium text-gray-700">{props.label}</span>
      {props.children}
    </label>
  )
}

export type InputFieldProps = {
  label: string
  value: string
  type?: string
  placeholder?: string
  error?: boolean
  onInput: JSX.EventHandlerUnion<HTMLInputElement, InputEvent>
}

export function InputField(props: InputFieldProps) {
  return (
    <FieldLabel label={props.label}>
      <input
        class={fieldBaseClasses(props.error)}
        type={props.type ?? 'text'}
        value={props.value}
        placeholder={props.placeholder}
        onInput={props.onInput}
      />
    </FieldLabel>
  )
}

export type SelectOption = {
  name: string
  value: string
}

export type SelectFieldProps = {
  label: string
  value: string
  options: SelectOption[]
  onChange: JSX.EventHandlerUnion<HTMLSelectElement, Event>
}

export function SelectField(props: SelectFieldProps) {
  return (
    <FieldLabel label={props.label}>
      <select class={fieldBaseClasses()} value={props.value} onChange={props.onChange}>
        {props.options.map((option) => (
          <option value={option.value}>{option.name}</option>
        ))}
      </select>
    </FieldLabel>
  )
}

export type TextareaFieldProps = {
  label: string
  value: string
  rows?: number
  onInput: JSX.EventHandlerUnion<HTMLTextAreaElement, InputEvent>
}

export function TextareaField(props: TextareaFieldProps) {
  return (
    <FieldLabel label={props.label}>
      <textarea
        class={fieldBaseClasses()}
        rows={props.rows ?? 6}
        value={props.value}
        onInput={props.onInput}
      />
    </FieldLabel>
  )
}

export type CheckboxFieldProps = {
  label: string
  checked: boolean
  onChange: JSX.EventHandlerUnion<HTMLInputElement, Event>
}

export function CheckboxField(props: CheckboxFieldProps) {
  return (
    <label class="flex items-center gap-3 text-sm text-gray-700">
      <input
        type="checkbox"
        class="size-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
        checked={props.checked}
        onChange={props.onChange}
      />
      <span>{props.label}</span>
    </label>
  )
}

export function SearchInputField(props: Omit<InputFieldProps, 'type'>) {
  return <InputField {...props} type="search" />
}

export function NumberInputField(props: Omit<InputFieldProps, 'type'>) {
  return <InputField {...props} type="number" />
}

export function PhoneInputField(props: Omit<InputFieldProps, 'type'>) {
  return <InputField {...props} type="tel" />
}

export function DateInputField(props: Omit<InputFieldProps, 'type'>) {
  return <InputField {...props} type="date" />
}

export function TimeInputField(props: Omit<InputFieldProps, 'type'>) {
  return <InputField {...props} type="time" />
}

export type FileInputFieldProps = {
  label: string
  accept?: string
  multiple?: boolean
  onChange: JSX.EventHandlerUnion<HTMLInputElement, Event>
}

export function FileInputField(props: FileInputFieldProps) {
  return (
    <FieldLabel label={props.label}>
      <input
        class={classes(
          fieldBaseClasses(),
          'cursor-pointer file:mr-4 file:rounded-lg file:border-0 file:bg-blue-50 file:px-3 file:py-2 file:text-sm file:font-medium file:text-blue-700 hover:file:bg-blue-100'
        )}
        type="file"
        accept={props.accept}
        multiple={props.multiple}
        onChange={props.onChange}
      />
    </FieldLabel>
  )
}

export type RadioOption = {
  label: string
  value: string
  hint?: string
}

export type RadioGroupFieldProps = {
  label: string
  name: string
  value: string
  options: RadioOption[]
  onChange: JSX.EventHandlerUnion<HTMLInputElement, Event>
}

export function RadioGroupField(props: RadioGroupFieldProps) {
  return (
    <FieldLabel label={props.label}>
      <div class="space-y-3 rounded-2xl border border-gray-200 bg-white p-4 shadow-sm">
        <For each={props.options}>
          {(option) => (
            <label class="flex items-start gap-3 text-sm text-gray-700">
              <input
                type="radio"
                name={props.name}
                value={option.value}
                checked={props.value === option.value}
                onChange={props.onChange}
                class="mt-0.5 size-4 border-gray-300 text-blue-600 focus:ring-blue-500"
              />
              <span class="space-y-0.5">
                <span class="block font-medium text-gray-900">{option.label}</span>
                <Show when={option.hint}>
                  <span class="block text-gray-500">{option.hint}</span>
                </Show>
              </span>
            </label>
          )}
        </For>
      </div>
    </FieldLabel>
  )
}

export type ToggleFieldProps = {
  label: string
  checked: boolean
  onChange: JSX.EventHandlerUnion<HTMLInputElement, Event>
}

export function ToggleField(props: ToggleFieldProps) {
  return (
    <label class="flex items-center justify-between gap-4 rounded-2xl border border-gray-200 bg-white px-4 py-3 shadow-sm">
      <span class="text-sm font-medium text-gray-700">{props.label}</span>
      <span class="relative inline-flex items-center">
        <input
          type="checkbox"
          class="peer sr-only"
          checked={props.checked}
          onChange={props.onChange}
        />
        <span class="h-6 w-11 rounded-full bg-gray-300 transition peer-checked:bg-blue-600" />
        <span class="pointer-events-none absolute left-0.5 top-0.5 size-5 rounded-full bg-white shadow transition peer-checked:translate-x-5" />
      </span>
    </label>
  )
}

export type RangeFieldProps = {
  label: string
  value: string | number
  min?: number
  max?: number
  step?: number
  onInput: JSX.EventHandlerUnion<HTMLInputElement, InputEvent>
}

export function RangeField(props: RangeFieldProps) {
  return (
    <FieldLabel label={props.label}>
      <input
        type="range"
        min={props.min}
        max={props.max}
        step={props.step}
        value={props.value}
        onInput={props.onInput}
        class="h-2 w-full cursor-pointer appearance-none rounded-full bg-gray-200 accent-blue-600"
      />
    </FieldLabel>
  )
}

export type FloatingLabelFieldProps = {
  label: string
  value: string
  type?: string
  onInput: JSX.EventHandlerUnion<HTMLInputElement, InputEvent>
}

export function FloatingLabelField(props: FloatingLabelFieldProps) {
  return (
    <label class="relative block">
      <input
        class="peer block w-full rounded-2xl border border-gray-300 bg-white px-3 pb-2 pt-6 text-sm text-gray-900 shadow-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-100"
        type={props.type ?? 'text'}
        value={props.value}
        placeholder=" "
        onInput={props.onInput}
      />
      <span class="pointer-events-none absolute left-3 top-2 text-xs font-medium uppercase tracking-wide text-gray-500 transition peer-placeholder-shown:top-4 peer-placeholder-shown:text-sm peer-placeholder-shown:normal-case peer-focus:top-2 peer-focus:text-xs peer-focus:uppercase peer-focus:tracking-wide">
        {props.label}
      </span>
    </label>
  )
}
