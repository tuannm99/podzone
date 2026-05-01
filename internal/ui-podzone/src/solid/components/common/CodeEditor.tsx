import { basicSetup, EditorView } from 'codemirror';
import { go } from '@codemirror/lang-go';
import { javascript } from '@codemirror/lang-javascript';
import { python } from '@codemirror/lang-python';
import { rust } from '@codemirror/lang-rust';
import { Vim, getCM, vim } from '@replit/codemirror-vim';
import { Compartment, EditorState, type Extension } from '@codemirror/state';
import { indentUnit } from '@codemirror/language';
import {
  Show,
  createEffect,
  createMemo,
  createSignal,
  onCleanup,
  onMount,
} from 'solid-js';
import { classes } from '../../shared/utils';

type VimMode = 'insert' | 'normal' | 'visual' | 'replace';

type CodeEditorProps = {
  label?: string;
  language?: string;
  value: string;
  rows?: number;
  class?: string;
  hint?: string;
  onInput: (value: string) => void;
};

type VimModeChange = {
  mode?: string;
  subMode?: string;
};

function resolveLanguageExtension(language: string | undefined): Extension {
  switch ((language ?? '').toLowerCase()) {
    case 'go':
      return go();
    case 'javascript':
      return javascript();
    case 'python':
      return python();
    case 'rust':
      return rust();
    case 'typescript':
      return javascript({ typescript: true });
    default:
      return [];
  }
}

function normalizeMode(mode: string | undefined): VimMode {
  if (mode === 'insert' || mode === 'visual' || mode === 'replace') {
    return mode;
  }
  return 'normal';
}

export function CodeEditor(props: CodeEditorProps) {
  let containerRef: HTMLDivElement | undefined;
  let view: EditorView | undefined;
  let applyingExternalValue = false;
  let vimListener: ((event: VimModeChange) => void) | undefined;
  const languageCompartment = new Compartment();

  const [mode, setMode] = createSignal<VimMode>('insert');
  const lineCount = createMemo(() =>
    props.value === '' ? 1 : props.value.split('\n').length
  );
  const minHeight = createMemo(
    () => `${Math.max(props.rows ?? 22, 16) * 24 + 32}px`
  );
  const modeHint = createMemo(() => {
    if (props.hint) return props.hint;
    if (mode() === 'insert') {
      return 'Insert mode. Esc switches to Vim normal mode.';
    }
    if (mode() === 'visual') {
      return 'Visual mode. Use h/j/k/l to move and y/d to act on the selection.';
    }
    if (mode() === 'replace') {
      return 'Replace mode. Characters overwrite existing content until Esc.';
    }
    return 'Normal mode. Press i to type, a to append, o to open a new line.';
  });

  onMount(() => {
    if (!containerRef) return;

    const editorTheme = EditorView.theme({
      '&': {
        'min-height': minHeight(),
        'border-radius': '1rem',
        border: '1px solid rgb(203 213 225)',
        background:
          'linear-gradient(180deg, rgb(248 250 252), rgb(255 255 255))',
        overflow: 'hidden',
      },
      '.cm-scroller': {
        'min-height': minHeight(),
        'font-family':
          '"JetBrains Mono", "SFMono-Regular", ui-monospace, "Cascadia Code", "Source Code Pro", Menlo, Monaco, Consolas, "Liberation Mono", monospace',
      },
      '.cm-content, .cm-gutter': {
        'font-family':
          '"JetBrains Mono", "SFMono-Regular", ui-monospace, "Cascadia Code", "Source Code Pro", Menlo, Monaco, Consolas, "Liberation Mono", monospace',
        'font-size': '13px',
        'line-height': '24px',
      },
      '.cm-content': {
        padding: '16px 0',
      },
      '.cm-line': {
        padding: '0 16px',
      },
      '.cm-gutters': {
        background: 'rgb(241 245 249 / 0.88)',
        color: 'rgb(100 116 139)',
        border: 'none',
        'border-right': '1px solid rgb(226 232 240)',
      },
      '.cm-activeLine': {
        background: 'rgb(219 234 254 / 0.4)',
      },
      '.cm-activeLineGutter': {
        background: 'rgb(219 234 254 / 0.9)',
        color: 'rgb(29 78 216)',
        'font-weight': '700',
      },
      '.cm-selectionBackground, &.cm-focused .cm-selectionBackground, ::selection':
        {
          background: 'rgb(191 219 254 / 0.75)',
        },
      '.cm-cursor, .cm-dropCursor': {
        'border-left-color': 'rgb(15 23 42)',
      },
      '&.cm-focused': {
        outline: '2px solid rgb(147 197 253)',
        'outline-offset': '0',
      },
    });

    view = new EditorView({
      state: EditorState.create({
        doc: props.value,
        extensions: [
          vim(),
          basicSetup,
          indentUnit.of('  '),
          languageCompartment.of(resolveLanguageExtension(props.language)),
          editorTheme,
          EditorView.updateListener.of((update) => {
            if (update.docChanged && !applyingExternalValue) {
              props.onInput(update.state.doc.toString());
            }
          }),
        ],
      }),
      parent: containerRef,
    });

    const cm = getCM(view);
    if (cm) {
      vimListener = (event: VimModeChange) => {
        setMode(normalizeMode(event.mode));
      };
      cm.on('vim-mode-change', vimListener);
      Vim.handleKey(cm, 'i', 'user');
      setMode('insert');
    }

    onCleanup(() => {
      const currentView = view;
      if (currentView && vimListener) {
        const currentCM = getCM(currentView);
        currentCM?.off('vim-mode-change', vimListener);
      }
      view?.destroy();
      view = undefined;
    });
  });

  createEffect(() => {
    const nextValue = props.value;
    const currentView = view;
    if (!currentView) return;

    const currentValue = currentView.state.doc.toString();
    if (currentValue === nextValue) return;

    applyingExternalValue = true;
    currentView.dispatch({
      changes: {
        from: 0,
        to: currentValue.length,
        insert: nextValue,
      },
    });
    applyingExternalValue = false;
  });

  createEffect(() => {
    const currentView = view;
    if (!currentView) return;

    currentView.dispatch({
      effects: languageCompartment.reconfigure(
        resolveLanguageExtension(props.language)
      ),
    });
  });

  createEffect(() => {
    const nextHeight = minHeight();
    const currentView = view;
    if (!currentView) return;

    currentView.dom.style.minHeight = nextHeight;
    const scroller = currentView.scrollDOM;
    scroller.style.minHeight = nextHeight;
  });

  return (
    <div class={classes('space-y-3', props.class)}>
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div class="space-y-1">
          <div class="text-sm font-medium text-gray-800">
            {props.label ?? 'Code'}
          </div>
          <div class="text-xs text-gray-500">{modeHint()}</div>
        </div>

        <div class="flex flex-wrap items-center gap-2 text-xs text-gray-500">
          <span class="rounded-full bg-blue-50 px-2.5 py-1 font-semibold uppercase tracking-wide text-blue-700">
            {props.language ?? 'code'}
          </span>
          <span
            class={classes(
              'rounded-full px-2.5 py-1 font-semibold uppercase tracking-wide',
              mode() === 'insert' && 'bg-emerald-50 text-emerald-700',
              mode() === 'normal' && 'bg-amber-50 text-amber-700',
              mode() === 'visual' && 'bg-fuchsia-50 text-fuchsia-700',
              mode() === 'replace' && 'bg-rose-50 text-rose-700'
            )}
          >
            {mode()}
          </span>
          <span>{lineCount()} lines</span>
        </div>
      </div>

      <div class="overflow-hidden rounded-2xl bg-white shadow-sm">
        <div class="flex items-center justify-between border border-slate-200 border-b-0 bg-white/90 px-4 py-2 text-xs text-slate-500">
          <span class="font-medium uppercase tracking-[0.18em] text-slate-600">
            Editor
          </span>
          <span>CodeMirror 6 + Vim</span>
        </div>

        <div ref={containerRef} />
      </div>

      <Show when={mode() === 'normal'}>
        <div class="rounded-xl border border-amber-200 bg-amber-50 px-3 py-2 text-xs text-amber-800">
          Normal mode is active. Press{' '}
          <kbd class="rounded bg-white px-1.5 py-0.5 font-mono">i</kbd> to type.
        </div>
      </Show>
    </div>
  );
}
