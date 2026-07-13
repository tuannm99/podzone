import { Component, input } from '@angular/core';

@Component({
  selector: 'app-section-title',
  templateUrl: './section-title.component.html',
})
export class SectionTitle {
  title = input.required<string>();
  subtitle = input<string>();
}
