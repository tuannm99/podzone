import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';

@Component({
  selector: 'app-panel-header',
  templateUrl: './panel-header.component.html',
})
export class PanelHeader {
  title = input.required<string>();
  copy = input<string>();
  class = input<string>();

  protected className = computed(() =>
    classes('flex flex-col gap-4 md:flex-row md:items-start md:justify-between', this.class()),
  );
}
