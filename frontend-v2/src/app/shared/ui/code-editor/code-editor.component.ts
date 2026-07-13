import { Component, computed, input, output, signal } from '@angular/core';
import { classes } from '../../utils';

// STUB: the Solid source (frontend/packages/shared/ui/components/common/CodeEditor.tsx)
// wraps CodeMirror 6 + @replit/codemirror-vim for real Vim-mode editing.
// None of @codemirror/*, codemirror, or @replit/codemirror-vim are installed
// in frontend-v2/package.json yet (adding a dependency is outside this
// port's scope — see PZEP-0006). This component ports the props/outputs
// and the static header UI faithfully, but renders a plain <textarea>
// instead of a real CodeMirror instance. `mode` is always 'insert' here
// since there's no real Vim integration. Wire up CodeMirror in the
// constructor/an afterNextRender effect once the dependency is added.

type VimMode = 'insert' | 'normal' | 'visual' | 'replace';

@Component({
  selector: 'app-code-editor',
  templateUrl: './code-editor.component.html',
})
export class CodeEditor {
  label = input<string>();
  language = input<string>();
  value = input.required<string>();
  rows = input(22);
  class = input<string>();
  hint = input<string>();
  valueChange = output<string>();

  protected mode = signal<VimMode>('insert');

  protected lineCount = computed(() => (this.value() === '' ? 1 : this.value().split('\n').length));
  protected minHeight = computed(() => `${Math.max(this.rows(), 16) * 24 + 32}px`);
  protected modeHint = computed(
    () => this.hint() ?? 'Insert mode. Esc switches to Vim normal mode.',
  );

  protected wrapperClass = computed(() => classes('space-y-3', this.class()));

  protected modeBadgeClass = computed(() =>
    classes(
      'rounded-full px-2.5 py-1 font-semibold uppercase tracking-wide',
      this.mode() === 'insert' && 'bg-emerald-50 text-emerald-700',
      this.mode() === 'normal' && 'bg-amber-50 text-amber-700',
      this.mode() === 'visual' && 'bg-fuchsia-50 text-fuchsia-700',
      this.mode() === 'replace' && 'bg-rose-50 text-rose-700',
    ),
  );

  protected onInput(value: string) {
    this.valueChange.emit(value);
  }
}
