import type { JSX } from 'solid-js'
import {
  CheckboxField,
  InputField,
  SelectField,
  TextareaField,
  type SelectOption,
} from '@/solid/components/common/Primitives'
import type { FormStore } from './createFormStore'

type StringKey<TValues> = Extract<keyof TValues, string>

type FormFieldProps<TValues extends Record<string, unknown>> = {
  form: FormStore<TValues>
  name: StringKey<TValues>
  label: string
}

function stringValue(value: unknown) {
  return typeof value === 'string' ? value : String(value ?? '')
}

export function FormInputField<TValues extends Record<string, unknown>>(
  props: FormFieldProps<TValues> & {
    type?: string
    placeholder?: string
    disabled?: boolean
    onValueInput?: (value: string) => void
  }
) {
  return (
    <InputField
      label={props.label}
      type={props.type}
      placeholder={props.placeholder}
      disabled={props.disabled}
      value={stringValue(props.form.value(props.name))}
      error={props.form.hasError(props.name)}
      errorText={props.form.error(props.name)}
      onInput={(event) => {
        const value = event.currentTarget.value
        props.form.setValue(props.name, value as TValues[typeof props.name])
        props.onValueInput?.(value)
      }}
    />
  )
}

export function FormSelectField<TValues extends Record<string, unknown>>(
  props: FormFieldProps<TValues> & {
    options: SelectOption[]
    onValueChange?: (value: string) => void
  }
) {
  return (
    <SelectField
      label={props.label}
      options={props.options}
      value={stringValue(props.form.value(props.name))}
      error={props.form.hasError(props.name)}
      errorText={props.form.error(props.name)}
      onChange={(event) => {
        const value = event.currentTarget.value
        props.form.setValue(props.name, value as TValues[typeof props.name])
        props.onValueChange?.(value)
      }}
    />
  )
}

export function FormTextareaField<TValues extends Record<string, unknown>>(
  props: FormFieldProps<TValues> & {
    rows?: number
  }
) {
  return (
    <TextareaField
      label={props.label}
      rows={props.rows}
      value={stringValue(props.form.value(props.name))}
      error={props.form.hasError(props.name)}
      errorText={props.form.error(props.name)}
      onInput={(event) =>
        props.form.setValue(
          props.name,
          event.currentTarget.value as TValues[typeof props.name]
        )
      }
    />
  )
}

export function FormCheckboxField<TValues extends Record<string, unknown>>(
  props: FormFieldProps<TValues>
) {
  return (
    <CheckboxField
      label={props.label}
      checked={Boolean(props.form.value(props.name))}
      onChange={(event) =>
        props.form.setValue(
          props.name,
          event.currentTarget.checked as TValues[typeof props.name]
        )
      }
    />
  )
}

export type FormSubmitHandler = JSX.EventHandler<HTMLFormElement, SubmitEvent>
