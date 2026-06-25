import { createSignal, type Setter } from 'solid-js';
import { createStore } from 'solid-js/store';
import type { ValidatorMap } from './validators';

type StringKey<TValues> = Extract<keyof TValues, string>;
type FormErrors<TValues> = Partial<Record<StringKey<TValues>, string>>;
type FormTouched<TValues> = Partial<Record<StringKey<TValues>, boolean>>;

type CreateFormStoreOptions<TValues extends Record<string, unknown>> = {
  initialValues: TValues;
  validators?: ValidatorMap<TValues>;
};

export type FormStore<TValues extends Record<string, unknown>> = {
  values: TValues;
  errors: FormErrors<TValues>;
  touched: FormTouched<TValues>;
  isSubmitting: () => boolean;
  value: <K extends StringKey<TValues>>(name: K) => TValues[K];
  error: <K extends StringKey<TValues>>(name: K) => string;
  hasError: <K extends StringKey<TValues>>(name: K) => boolean;
  setValue: <K extends StringKey<TValues>>(name: K, value: TValues[K]) => void;
  setSubmitting: Setter<boolean>;
  validate: () => boolean;
  validateField: <K extends StringKey<TValues>>(name: K) => boolean;
  reset: (next?: Partial<TValues>) => void;
};

export function createFormStore<TValues extends Record<string, unknown>>(
  options: CreateFormStoreOptions<TValues>
): FormStore<TValues> {
  const [values, setValues] = createStore<Record<string, unknown>>({
    ...options.initialValues,
  });
  const [errors, setErrors] = createStore<Record<string, string | undefined>>({});
  const [touched, setTouched] = createStore<Record<string, boolean>>({});
  const [submitting, setSubmitting] = createSignal(false);

  const typedValues = values as TValues;

  const validateField = <K extends StringKey<TValues>>(name: K) => {
    const validators = options.validators?.[name] || [];
    const message = validators
      .map((validator) => validator(typedValues[name], typedValues))
      .find(Boolean);
    setErrors(name, message || undefined);
    return !message;
  };

  const validate = () => {
    const fieldNames = Object.keys(options.initialValues) as StringKey<TValues>[];
    let valid = true;
    for (const fieldName of fieldNames) {
      setTouched(fieldName, true);
      valid = validateField(fieldName) && valid;
    }
    return valid;
  };

  const setValue = <K extends StringKey<TValues>>(name: K, value: TValues[K]) => {
    setValues(name, value);
    setTouched(name, true);
    if (touched[name] || errors[name]) {
      validateField(name);
    }
  };

  const reset = (next?: Partial<TValues>) => {
    setValues({ ...options.initialValues, ...next });
    setErrors({});
    setTouched({});
    setSubmitting(false);
  };

  return {
    values: typedValues,
    errors: errors as FormErrors<TValues>,
    touched: touched as FormTouched<TValues>,
    isSubmitting: submitting,
    value: (name) => typedValues[name],
    error: (name) => errors[name] || '',
    hasError: (name) => Boolean(touched[name] && errors[name]),
    setValue,
    setSubmitting,
    validate,
    validateField,
    reset,
  };
}
