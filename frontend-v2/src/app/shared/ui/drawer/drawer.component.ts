import { Component, computed, input, output } from '@angular/core';
import { classes, createUniqueId } from '../../utils';

export type DrawerSide = 'left' | 'right';

const drawerSideClasses: Record<DrawerSide, string> = {
  left: 'left-0',
  right: 'right-0',
};

@Component({
  selector: 'app-drawer',
  templateUrl: './drawer.component.html',
})
export class Drawer {
  open = input.required<boolean>();
  title = input<string>();
  side = input<DrawerSide>('right');
  class = input<string>();
  closed = output<void>();

  // Same focus-trap deferral as Modal — see modal.component.ts's note.
  protected headingId = createUniqueId();

  protected panelClassName = computed(() =>
    classes(
      'absolute top-0 h-full w-full max-w-md overflow-y-auto border-gray-200 bg-white p-6 shadow-xl',
      drawerSideClasses[this.side()],
      this.side() === 'left' ? 'border-r' : 'border-l',
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
