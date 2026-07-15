import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';

// Tone names match the semantic outcomes a status pill needs to convey
// (e.g. onboarding store `ui_state`: ready/failed/pending|provisioning|
// blocked), not visual colors directly — see badge.component.scss for the
// actual color mapping per tone.
export type BadgeTone = 'success' | 'danger' | 'warning' | 'neutral';

@Component({
  selector: 'app-badge',
  template: `<span class="badge" [class]="toneClass()"><ng-content /></span>`,
  styleUrl: './badge.component.scss',
})
export class Badge {
  tone = input<BadgeTone>('neutral');

  protected toneClass = computed(() => classes(`badge--${this.tone()}`));
}
