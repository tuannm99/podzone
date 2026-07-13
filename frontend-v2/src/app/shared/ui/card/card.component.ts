import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';

@Component({
  selector: 'app-card',
  templateUrl: './card.component.html',
})
export class Card {
  class = input<string>();

  protected className = computed(() =>
    classes('rounded-lg border border-gray-200 bg-white p-5 shadow-sm', this.class()),
  );
}
