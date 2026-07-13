import { Component, computed, input, output } from '@angular/core';
import { classes, createUniqueId } from '../../utils';

export type ModalSize = 'sm' | 'md' | 'lg' | 'xl';

const modalSizeClasses: Record<ModalSize, string> = {
  sm: 'max-w-md',
  md: 'max-w-2xl',
  lg: 'max-w-4xl',
  xl: 'max-w-6xl',
};

@Component({
  selector: 'app-modal',
  templateUrl: './modal.component.html',
})
export class Modal {
  open = input.required<boolean>();
  title = input<string>();
  size = input<ModalSize>('md');
  class = input<string>();
  closed = output<void>();

  // NOTE: real focus-trap/focus-restore behavior (Tab/Shift-Tab trapped,
  // focus moved in on open and restored to the trigger on close) is NOT
  // implemented here — the Solid original uses a `useFocusTrap` hook with
  // no Angular equivalent yet. Per agent/ANGULAR_STYLE_GUIDE.md this needs
  // @angular/cdk/a11y, which is not yet a dependency of this project.
  // role/aria-modal/aria-labelledby/Escape/click-outside are all wired;
  // focus management is the deferred piece.
  protected headingId = createUniqueId();

  protected panelClassName = computed(() =>
    classes(
      'w-full rounded-lg border border-gray-200 bg-white p-5 shadow-xl',
      modalSizeClasses[this.size()],
      this.class(),
    ),
  );

  protected onBackdropClick(event: MouseEvent) {
    if (event.target === event.currentTarget) {
      this.closed.emit();
    }
  }

  protected onEscape() {
    this.closed.emit();
  }
}
