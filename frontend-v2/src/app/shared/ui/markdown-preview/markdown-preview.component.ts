import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';

// STUB: the Solid source wraps @toast-ui/editor's Viewer for real rendered
// Markdown. @toast-ui/editor is not installed in frontend-v2/package.json
// yet (adding it is outside this port's scope — see PZEP-0006). This
// renders the raw Markdown source in a <pre> instead of rendered HTML.
// Wire up the real Toast UI Viewer once the dependency is added.

@Component({
  selector: 'app-markdown-preview',
  templateUrl: './markdown-preview.component.html',
})
export class MarkdownPreview {
  source = input.required<string>();
  class = input<string>();
  emptyMessage = input<string>();

  protected hasSource = computed(() => this.source().trim().length > 0);

  protected emptyClass = computed(() =>
    classes(
      'rounded-lg border border-dashed border-slate-300 bg-slate-50 px-6 py-8 text-sm leading-7 text-slate-500',
      this.class(),
    ),
  );

  protected viewerClass = computed(() =>
    classes('rounded-lg border border-slate-200 bg-white shadow-sm', this.class()),
  );
}
