import { Component, computed, input, output } from '@angular/core';
import { classes } from '../../utils';

// STUB: the Solid source wraps @toast-ui/editor's full WYSIWYG Editor.
// @toast-ui/editor is not installed in frontend-v2/package.json yet
// (adding it is outside this port's scope — see PZEP-0006). This renders
// a plain <textarea> instead of the real WYSIWYG/Markdown-toggle editor.
// Wire up the real Toast UI Editor once the dependency is added.

@Component({
  selector: 'app-rich-text-editor',
  templateUrl: './rich-text-editor.component.html',
})
export class RichTextEditor {
  label = input<string>();
  value = input.required<string>();
  class = input<string>();
  hint = input<string>();
  placeholder = input<string>();
  valueChange = output<string>();

  protected wrapperClass = computed(() => classes('space-y-3', this.class()));

  protected onInput(value: string) {
    this.valueChange.emit(value);
  }
}
