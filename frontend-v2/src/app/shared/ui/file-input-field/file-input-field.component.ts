import { Component, computed, input, output } from '@angular/core';
import { FieldLabel } from '../field-label/field-label.component';
import { fieldBaseClasses } from '../field-classes';
import { classes, createUniqueId } from '../../utils';

@Component({
  selector: 'app-file-input-field',
  imports: [FieldLabel],
  templateUrl: './file-input-field.component.html',
})
export class FileInputField {
  label = input.required<string>();
  accept = input<string>();
  multiple = input(false);

  fileChange = output<FileList | null>();

  protected id = createUniqueId();
  protected fieldClass = computed(() =>
    classes(
      fieldBaseClasses(),
      'cursor-pointer file:mr-4 file:rounded-md file:border-0 file:bg-gray-100 file:px-3 file:py-2 file:text-sm file:font-medium file:text-gray-800 hover:file:bg-gray-200',
    ),
  );
}
