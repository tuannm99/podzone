import { Component, computed, input, output } from '@angular/core';
import { classes } from '../../utils';
import { type AlertTone, toneClasses } from '../display-shared';

// NOTE (port judgment call): Solid's close button visibility is driven by
// whether an `onClose` callback prop was passed at all
// (`<Show when={props.onClose}>`). Angular can't inspect "is anyone
// listening to this output" the same way, so this exposes an explicit
// `closable` input (default true) instead — caller sets `[closable]=false`
// to hide the button rather than simply not binding `(close)`.
@Component({
  selector: 'app-toast',
  templateUrl: './toast.component.html',
})
export class Toast {
  show = input(true);
  tone = input<AlertTone>('dark');
  title = input<string>();
  fixed = input(true);
  closable = input(true);
  class = input<string>();

  close = output<void>();

  protected wrapperClass = computed(() =>
    classes(
      'z-50 flex max-w-sm items-start justify-between gap-4 rounded-lg border bg-white px-4 py-3 shadow-xl',
      this.fixed() && 'fixed bottom-4 right-4',
      toneClasses[this.tone()],
      this.class(),
    ),
  );
}
