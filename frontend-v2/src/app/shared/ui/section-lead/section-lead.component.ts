import { Component, input } from '@angular/core';

@Component({
  selector: 'app-section-lead',
  templateUrl: './section-lead.component.html',
  styleUrl: './section-lead.component.scss',
})
export class SectionLead {
  eyebrow = input.required<string>();
  title = input.required<string>();
  copy = input.required<string>();
}
