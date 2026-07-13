import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';

@Component({
  selector: 'app-button-group',
  templateUrl: './button-group.component.html',
})
export class ButtonGroup {
  class = input<string>();

  protected className = computed(() =>
    classes('inline-flex flex-wrap items-center gap-2', this.class()),
  );
}
