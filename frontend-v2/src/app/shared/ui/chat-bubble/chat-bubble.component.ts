import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';

@Component({
  selector: 'app-chat-bubble',
  templateUrl: './chat-bubble.component.html',
})
export class ChatBubble {
  author = input<string>();
  copy = input.required<string>();
  meta = input<string>();
  align = input<'start' | 'end'>('start');
  class = input<string>();

  protected isEnd = computed(() => this.align() === 'end');

  protected wrapperClass = computed(() =>
    classes('flex w-full', this.isEnd() ? 'justify-end' : 'justify-start', this.class()),
  );

  protected bubbleClass = computed(() =>
    classes(
      'max-w-xl rounded-lg px-4 py-3 shadow-sm',
      this.isEnd() ? 'bg-gray-950 text-white' : 'bg-white text-gray-900 ring-1 ring-gray-200',
    ),
  );

  protected metaClass = computed(() =>
    classes(
      'mb-1 flex flex-wrap items-center gap-2 text-xs',
      this.isEnd() ? 'text-gray-300' : 'text-gray-500',
    ),
  );
}
