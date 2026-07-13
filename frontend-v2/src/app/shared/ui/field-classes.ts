// Ported from the module-private `fieldBaseClasses` helper in
// frontend/packages/shared/ui/components/common/Primitives.tsx — shared
// across InputField/SelectField/TextareaField/FileInputField here since
// Angular files can't have Solid's module-private-function-per-file shape
// without duplicating this string in every component.
import { classes } from '../utils';

export function fieldBaseClasses(hasError?: boolean) {
  return classes(
    'block h-10 w-full rounded-md border bg-white px-3 text-sm text-gray-900 outline-none transition',
    hasError
      ? 'border-red-300 focus:border-red-500 focus:ring-2 focus:ring-red-100'
      : 'border-gray-300 focus:border-gray-950 focus:ring-2 focus:ring-gray-100',
  );
}
