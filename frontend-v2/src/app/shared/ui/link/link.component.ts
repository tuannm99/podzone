import { Component, computed, input } from '@angular/core';
import { RouterLink } from '@angular/router';
import { isExternalUrl } from '../../utils';

@Component({
  selector: 'app-link',
  imports: [RouterLink],
  templateUrl: './link.component.html',
})
export class Link {
  href = input.required<string>();
  target = input<string>();
  class = input<string>();

  protected isExternal = computed(() => isExternalUrl(this.href()));
  protected rel = computed(() => (this.target() === '_blank' ? 'noopener noreferrer' : undefined));
}
