import Viewer from '@toast-ui/editor/viewer';
import type ToastViewer from '@toast-ui/editor/viewer';
import '@toast-ui/editor/dist/toastui-editor-viewer.css';
import { Show, createEffect, onCleanup, onMount } from 'solid-js';
import { classes } from '../../shared/utils';

export function MarkdownPreview(props: {
  source: string;
  class?: string;
  emptyMessage?: string;
}) {
  let containerRef: HTMLDivElement | undefined;
  let viewer: ToastViewer | undefined;
  let lastSource = props.source;

  onMount(() => {
    if (!containerRef) return;

    viewer = new Viewer({
      el: containerRef,
      initialValue: props.source,
      usageStatistics: false,
      linkAttributes: {
        target: '_blank',
        rel: 'noreferrer noopener',
      },
    });

    lastSource = props.source;

    onCleanup(() => {
      viewer?.destroy();
      viewer = undefined;
    });
  });

  createEffect(() => {
    const nextSource = props.source;
    const currentViewer = viewer;
    if (!currentViewer || nextSource === lastSource) return;

    currentViewer.setMarkdown(nextSource);
    lastSource = nextSource;
  });

  return (
    <Show
      when={props.source.trim()}
      fallback={
        <div
          class={classes(
            'rounded-[28px] border border-dashed border-slate-300 bg-slate-50 px-6 py-8 text-sm leading-7 text-slate-500',
            props.class
          )}
        >
          {props.emptyMessage ?? 'No Markdown content yet.'}
        </div>
      }
    >
      <div
        class={classes(
          'problem-statement-viewer problem-statement-surface rounded-[28px] border border-slate-200 bg-gradient-to-b from-slate-50 via-white to-white shadow-sm',
          props.class
        )}
      >
        <div ref={containerRef} />
      </div>
    </Show>
  );
}
