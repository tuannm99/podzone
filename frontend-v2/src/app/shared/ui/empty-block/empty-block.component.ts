import { Component, input } from '@angular/core';

@Component({
  selector: 'app-empty-block',
  templateUrl: './empty-block.component.html',
  styleUrl: './empty-block.component.scss',
})
export class EmptyBlock {
  title = input.required<string>();
  copy = input.required<string>();
}
