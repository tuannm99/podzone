import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';

export type OverlayPosition = 'top' | 'right' | 'bottom' | 'left';

const floatingPositionClasses: Record<OverlayPosition, string> = {
  top: 'bottom-full left-1/2 mb-2 -translate-x-1/2',
  right: 'left-full top-1/2 ml-2 -translate-y-1/2',
  bottom: 'left-1/2 top-full mt-2 -translate-x-1/2',
  left: 'right-full top-1/2 mr-2 -translate-y-1/2',
};

@Component({
  selector: 'app-tooltip',
  templateUrl: './tooltip.component.html',
})
export class Tooltip {
  position = input<OverlayPosition>('top');
  class = input<string>();
  panelClass = input<string>();

  protected containerClass = computed(() => classes('group relative inline-flex', this.class()));
  protected panelClassName = computed(() =>
    classes(
      'pointer-events-none absolute z-30 rounded-lg bg-gray-900 px-2.5 py-1.5 text-xs text-white opacity-0 shadow-lg transition group-hover:opacity-100',
      floatingPositionClasses[this.position()],
      this.panelClass(),
    ),
  );
}
