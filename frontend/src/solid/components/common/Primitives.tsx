import { For, Show, createUniqueId, splitProps, type JSX, type ParentProps } from 'solid-js'
import { classes } from '../../shared/utils'
import { Link } from './Link'

type ButtonColor = 'primary' | 'alternative' | 'light' | 'dark' | 'green' | 'red'
type ButtonSize = 'xs' | 'sm' | 'md'
type BadgeColor = 'blue' | 'indigo' | 'green' | 'yellow' | 'pink' | 'dark' | 'red'

const buttonColorClasses: Record<ButtonColor, string> = {
    primary: 'bg-gray-950 text-white hover:bg-gray-800 focus:ring-gray-300',
    alternative: 'border border-gray-300 bg-white text-gray-900 hover:bg-gray-50 focus:ring-gray-200',
    light: 'border border-gray-200 bg-white text-gray-700 hover:bg-gray-50 focus:ring-gray-200',
    dark: 'bg-gray-950 text-white hover:bg-gray-800 focus:ring-gray-300',
    green: 'bg-green-700 text-white hover:bg-green-800 focus:ring-green-300',
    red: 'bg-red-700 text-white hover:bg-red-800 focus:ring-red-300',
}

const buttonSizeClasses: Record<ButtonSize, string> = {
    xs: 'h-8 px-3 text-xs',
    sm: 'h-9 px-3 text-sm',
    md: 'h-10 px-4 text-sm',
}

const badgeColorClasses: Record<BadgeColor, string> = {
    blue: 'bg-blue-100 text-blue-800',
    indigo: 'bg-indigo-100 text-indigo-800',
    green: 'bg-green-100 text-green-800',
    yellow: 'bg-yellow-100 text-yellow-800',
    pink: 'bg-pink-100 text-pink-800',
    dark: 'bg-gray-100 text-gray-800',
    red: 'bg-red-100 text-red-800',
}

function fieldBaseClasses(hasError?: boolean) {
    return classes(
        'block h-10 w-full rounded-md border bg-white px-3 text-sm text-gray-900 outline-none transition',
        hasError
            ? 'border-red-300 focus:border-red-500 focus:ring-2 focus:ring-red-100'
            : 'border-gray-300 focus:border-gray-950 focus:ring-2 focus:ring-gray-100'
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
        <section class={classes('rounded-lg border border-gray-200 bg-white p-5 shadow-sm', props.class)}>
            {props.children}
        </section>
    )
}

export function Badge(props: { content: string; color?: BadgeColor; class?: string }) {
    return (
        <span
            class={classes(
                'inline-flex items-center rounded-md px-2 py-1 text-xs font-semibold',
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
        'onClick',
    ])

    const className = () =>
        classes(
            'inline-flex items-center justify-center gap-2 whitespace-nowrap font-medium focus:outline-none focus:ring-2 disabled:pointer-events-none disabled:opacity-60',
            buttonColorClasses[local.color ?? 'primary'],
            buttonSizeClasses[local.size ?? 'md'],
            local.pill ? 'rounded-full' : 'rounded-md',
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
        <Link
            href={local.href}
            target={local.target}
            class={className()}
            aria-disabled={local.disabled || local.loading}
            onClick={
                local.disabled || local.loading
                    ? (event: MouseEvent) => event.preventDefault()
                    : (local.onClick as JSX.EventHandlerUnion<HTMLAnchorElement, MouseEvent>)
            }
        >
            {content}
        </Link>
    ) : (
        <button
            type={local.type ?? 'button'}
            class={className()}
            disabled={local.disabled || local.loading}
            onClick={local.onClick as JSX.EventHandlerUnion<HTMLButtonElement, MouseEvent>}
            {...rest}
        >
            {content}
        </button>
    )
}

export function FieldLabel(props: ParentProps<{ label: string; for?: string; class?: string }>) {
    return (
        <div class={classes('space-y-1.5', props.class)}>
            <label for={props.for} class="block text-xs font-semibold uppercase text-gray-500">
                {props.label}
            </label>
            {props.children}
        </div>
    )
}

export type InputFieldProps = {
    label: string
    value: string
    id?: string
    type?: string
    placeholder?: string
    disabled?: boolean
    error?: boolean
    errorText?: string
    onInput: JSX.EventHandlerUnion<HTMLInputElement, InputEvent>
}

export function InputField(props: InputFieldProps) {
    const uid = props.id ?? createUniqueId()
    const errorId = () => (props.errorText ? `${uid}-error` : undefined)
    return (
        <FieldLabel label={props.label} for={uid}>
            <input
                id={uid}
                class={fieldBaseClasses(props.error)}
                type={props.type ?? 'text'}
                value={props.value}
                placeholder={props.placeholder}
                disabled={props.disabled}
                aria-invalid={props.error || undefined}
                aria-describedby={errorId()}
                onInput={props.onInput}
            />
            <Show when={props.errorText}>
                <span id={errorId()} class="block text-xs font-medium text-red-600">
                    {props.errorText}
                </span>
            </Show>
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
    id?: string
    options: SelectOption[]
    disabled?: boolean
    error?: boolean
    errorText?: string
    onChange: JSX.EventHandlerUnion<HTMLSelectElement, Event>
}

export function SelectField(props: SelectFieldProps) {
    const uid = props.id ?? createUniqueId()
    const errorId = () => (props.errorText ? `${uid}-error` : undefined)
    return (
        <FieldLabel label={props.label} for={uid}>
            <select
                id={uid}
                class={fieldBaseClasses(props.error)}
                value={props.value}
                disabled={props.disabled}
                aria-invalid={props.error || undefined}
                aria-describedby={errorId()}
                onChange={props.onChange}
            >
                <For each={props.options}>{(option) => <option value={option.value}>{option.name}</option>}</For>
            </select>
            <Show when={props.errorText}>
                <span id={errorId()} class="block text-xs font-medium text-red-600">
                    {props.errorText}
                </span>
            </Show>
        </FieldLabel>
    )
}

export type TextareaFieldProps = {
    label: string
    value: string
    id?: string
    rows?: number
    error?: boolean
    errorText?: string
    onInput: JSX.EventHandlerUnion<HTMLTextAreaElement, InputEvent>
}

export function TextareaField(props: TextareaFieldProps) {
    const uid = props.id ?? createUniqueId()
    const errorId = () => (props.errorText ? `${uid}-error` : undefined)
    return (
        <FieldLabel label={props.label} for={uid}>
            <textarea
                id={uid}
                class={classes(fieldBaseClasses(props.error), 'h-auto py-2.5')}
                rows={props.rows ?? 6}
                value={props.value}
                aria-invalid={props.error || undefined}
                aria-describedby={errorId()}
                onInput={props.onInput}
            />
            <Show when={props.errorText}>
                <span id={errorId()} class="block text-xs font-medium text-red-600">
                    {props.errorText}
                </span>
            </Show>
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
                class="size-4 rounded border-gray-300 text-gray-950 focus:ring-gray-300"
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
    const uid = createUniqueId()
    return (
        <FieldLabel label={props.label} for={uid}>
            <input
                id={uid}
                class={classes(
                    fieldBaseClasses(),
                    'cursor-pointer file:mr-4 file:rounded-md file:border-0 file:bg-gray-100 file:px-3 file:py-2 file:text-sm file:font-medium file:text-gray-800 hover:file:bg-gray-200'
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
        <fieldset class="space-y-1.5">
            <legend class="block text-xs font-semibold uppercase text-gray-500">{props.label}</legend>
            <div class="space-y-3 rounded-lg border border-gray-200 bg-white p-4 shadow-sm">
                <For each={props.options}>
                    {(option) => (
                        <label class="flex items-start gap-3 text-sm text-gray-700">
                            <input
                                type="radio"
                                name={props.name}
                                value={option.value}
                                checked={props.value === option.value}
                                onChange={props.onChange}
                                class="mt-0.5 size-4 border-gray-300 text-gray-950 focus:ring-gray-300"
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
        </fieldset>
    )
}

export type ToggleFieldProps = {
    label: string
    checked: boolean
    onChange: JSX.EventHandlerUnion<HTMLInputElement, Event>
}

export function ToggleField(props: ToggleFieldProps) {
    return (
        <label class="flex items-center justify-between gap-4 rounded-lg border border-gray-200 bg-white px-4 py-3 shadow-sm">
            <span class="text-sm font-medium text-gray-700">{props.label}</span>
            <span class="relative inline-flex items-center">
                <input type="checkbox" class="peer sr-only" checked={props.checked} onChange={props.onChange} />
                <span class="h-6 w-11 rounded-full bg-gray-300 transition peer-checked:bg-gray-950" />
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
    const uid = createUniqueId()
    return (
        <FieldLabel label={props.label} for={uid}>
            <input
                id={uid}
                type="range"
                min={props.min}
                max={props.max}
                step={props.step}
                value={props.value}
                onInput={props.onInput}
                class="h-2 w-full cursor-pointer appearance-none rounded-full bg-gray-200 accent-gray-950"
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
                class="peer block w-full rounded-md border border-gray-300 bg-white px-3 pb-2 pt-6 text-sm text-gray-900 outline-none transition focus:border-gray-950 focus:ring-2 focus:ring-gray-100"
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
