import Editor from '@toast-ui/editor';
import type { Editor as ToastEditor, PreviewStyle } from '@toast-ui/editor';
import '@toast-ui/editor/dist/toastui-editor.css';
import { createEffect, onCleanup, onMount } from 'solid-js';
import { classes } from '../../shared/utils';

type RichTextEditorProps = {
  label?: string;
  value: string;
  class?: string;
  hint?: string;
  height?: string;
  minHeight?: string;
  placeholder?: string;
  onInput: (value: string) => void;
};

const TOOLBAR_ITEMS = [
  ['heading', 'bold', 'italic', 'strike'],
  ['hr', 'quote'],
  ['ul', 'ol', 'task'],
  ['table', 'link'],
  ['code', 'codeblock'],
];

function resolvePreviewStyle(): PreviewStyle {
  if (typeof window === 'undefined') return 'tab';
  return window.innerWidth >= 1280 ? 'vertical' : 'tab';
}

export function RichTextEditor(props: RichTextEditorProps) {
  let containerRef: HTMLDivElement | undefined;
  let editor: ToastEditor | undefined;
  let syncingExternalValue = false;

  const syncPreviewStyle = () => {
    const currentEditor = editor;
    if (!currentEditor) return;

    const nextStyle = resolvePreviewStyle();
    if (currentEditor.getCurrentPreviewStyle() !== nextStyle) {
      currentEditor.changePreviewStyle(nextStyle);
    }
  };

  onMount(() => {
    if (!containerRef) return;

    editor = new Editor({
      el: containerRef,
      height: props.height ?? '620px',
      minHeight: props.minHeight ?? '460px',
      initialValue: props.value,
      initialEditType: 'wysiwyg',
      previewStyle: resolvePreviewStyle(),
      placeholder: props.placeholder,
      hideModeSwitch: false,
      usageStatistics: false,
      linkAttributes: {
        target: '_blank',
        rel: 'noreferrer noopener',
      },
      toolbarItems: TOOLBAR_ITEMS,
      events: {
        change: () => {
          const currentEditor = editor;
          if (!currentEditor || syncingExternalValue) return;
          props.onInput(currentEditor.getMarkdown());
        },
      },
    });

    syncPreviewStyle();
    window.addEventListener('resize', syncPreviewStyle);

    onCleanup(() => {
      window.removeEventListener('resize', syncPreviewStyle);
      editor?.destroy();
      editor = undefined;
    });
  });

  createEffect(() => {
    const nextValue = props.value;
    const currentEditor = editor;
    if (!currentEditor) return;

    if (currentEditor.getMarkdown() === nextValue) return;

    syncingExternalValue = true;
    currentEditor.setMarkdown(nextValue, false);
    syncingExternalValue = false;
  });

  return (
    <div class={classes('space-y-3', props.class)}>
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div class="space-y-1">
          <div class="text-sm font-medium text-gray-800">
            {props.label ?? 'Statement editor'}
          </div>
          <div class="text-xs text-gray-500">
            {props.hint ??
              'Write in WYSIWYG or Markdown. The saved source remains Markdown for portability.'}
          </div>
        </div>

        <div class="rounded-full bg-blue-50 px-3 py-1 text-xs font-semibold uppercase tracking-[0.18em] text-blue-700">
          TOAST UI Editor
        </div>
      </div>

      <div class="problem-statement-editor problem-statement-surface">
        <div ref={containerRef} />
      </div>
    </div>
  );
}
