import { NgTemplateOutlet } from '@angular/common';
import { Component, computed, input, output } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { RouterLink } from '@angular/router';
import { classes, isExternalUrl } from '../../utils';
import { Spinner } from '../spinner/spinner.component';

export type ButtonVariant = 'flat' | 'stroked';
export type ButtonColor = 'primary' | 'accent' | 'warn';

@Component({
  selector: 'app-button',
  imports: [RouterLink, Spinner, NgTemplateOutlet, MatButtonModule],
  templateUrl: './button.component.html',
  styleUrl: './button.component.scss',
})
export class Button {
  variant = input<ButtonVariant>('flat');
  color = input<ButtonColor>('primary');
  href = input<string>();
  target = input<string>();
  type = input<'button' | 'submit' | 'reset'>('button');
  loading = input(false);
  disabled = input(false);
  class = input<string>();
  // Angular doesn't forward a plain `aria-label="..."` attribute past a
  // component's host tag to its real inner <a>/<button> — the host tag
  // itself isn't the interactive element, so a static attribute there is
  // invisible to assistive tech reading the inner control. Consumers must
  // bind `[ariaLabel]`, not rely on a bare `aria-label` attribute.
  ariaLabel = input<string>();

  buttonClick = output<MouseEvent>();

  protected isExternal = computed(() => {
    const value = this.href();
    return value ? isExternalUrl(value) : false;
  });

  protected className = computed(() => classes('app-button', this.class()));

  protected onLinkClick(event: MouseEvent) {
    if (this.disabled() || this.loading()) {
      event.preventDefault();
      return;
    }
    this.buttonClick.emit(event);
  }
}
