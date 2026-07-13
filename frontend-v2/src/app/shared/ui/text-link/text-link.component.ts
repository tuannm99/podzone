import { Component, computed, input } from '@angular/core';
import { RouterLink } from '@angular/router';
import { classes, isExternalUrl } from '../../utils';

// Inlines the same internal-vs-external branching as Button
// (shared/ui/button) rather than depending on Link.tsx's Angular port,
// which is being ported by a different batch in parallel — avoids a
// cross-batch dependency on an unverified API during this pass.
@Component({
  selector: 'app-text-link',
  imports: [RouterLink],
  templateUrl: './text-link.component.html',
})
export class TextLink {
  href = input.required<string>();
  target = input<string>();
  class = input<string>();

  protected isExternal = computed(() => isExternalUrl(this.href()));

  protected className = computed(() =>
    classes(
      'font-medium text-gray-900 underline decoration-gray-300 underline-offset-4 transition hover:text-gray-700 hover:decoration-gray-500',
      this.class(),
    ),
  );

  protected rel = computed(() => (this.target() === '_blank' ? 'noreferrer' : undefined));
}
