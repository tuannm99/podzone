import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';
// NOTE: depends on Link.tsx's Angular port, assigned to a different
// parallel porting batch. Path assumed per this project's folder-per-
// component convention — reconcile at the consolidated build pass if the
// actual path differs.
import { Link } from '../link/link.component';

export type SpeedDialItem = {
  label: string;
  href?: string;
  onClick?: () => void;
};

@Component({
  selector: 'app-speed-dial',
  imports: [Link],
  templateUrl: './speed-dial.component.html',
})
export class SpeedDial {
  items = input.required<SpeedDialItem[]>();
  class = input<string>();

  protected className = computed(() =>
    classes('fixed bottom-6 right-6 z-40 hidden flex-col gap-2 md:flex', this.class()),
  );

  protected onItemClick(item: SpeedDialItem) {
    item.onClick?.();
  }
}
