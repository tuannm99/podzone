import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';
import type { OverlayPosition } from '../tooltip/tooltip.component';

const floatingPositionClasses: Record<OverlayPosition, string> = {
  top: 'bottom-full left-1/2 mb-2 -translate-x-1/2',
  right: 'left-full top-1/2 ml-2 -translate-y-1/2',
  bottom: 'left-1/2 top-full mt-2 -translate-x-1/2',
  left: 'right-full top-1/2 mr-2 -translate-y-1/2',
};

@Component({
  selector: 'app-popover',
  templateUrl: './popover.component.html',
})
export class Popover {
  position = input<OverlayPosition>('bottom');
  class = input<string>();
  panelClass = input<string>();

  protected containerClass = computed(() => classes('group relative inline-flex', this.class()));
  protected panelClassName = computed(() =>
    classes(
      'pointer-events-none absolute z-30 min-w-56 rounded-lg border border-gray-200 bg-white p-4 text-sm text-gray-600 opacity-0 shadow-xl transition group-hover:pointer-events-auto group-hover:opacity-100',
      floatingPositionClasses[this.position()],
      this.panelClass(),
    ),
  );
}
