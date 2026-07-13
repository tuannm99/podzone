import { Component, input } from '@angular/core';
import { Card } from '../card/card.component';

@Component({
  selector: 'app-stat-card',
  imports: [Card],
  templateUrl: './stat-card.component.html',
})
export class StatCard {
  label = input.required<string>();
  value = input.required<string>();
}
