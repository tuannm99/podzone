import { Component, computed, input } from '@angular/core';
import { MatCardModule } from '@angular/material/card';
import { classes } from '../../utils';

export type CardTone = 'surface' | 'inverse';

@Component({
  selector: 'app-card',
  imports: [MatCardModule],
  templateUrl: './card.component.html',
  styleUrl: './card.component.scss',
})
export class Card {
  tone = input<CardTone>('surface');
  class = input<string>();

  protected className = computed(() =>
    classes('app-card', `app-card--${this.tone()}`, this.class()),
  );
}
