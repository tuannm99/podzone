import { Component, computed, input, output } from '@angular/core';
import { RouterLink } from '@angular/router';
import { classes, isExternalUrl } from '../../utils';
import { Spinner } from '../spinner/spinner.component';

export type ButtonColor = 'primary' | 'alternative' | 'light' | 'dark' | 'green' | 'red';
export type ButtonSize = 'xs' | 'sm' | 'md';

const buttonColorClasses: Record<ButtonColor, string> = {
  primary: 'bg-gray-950 text-white hover:bg-gray-800 focus:ring-gray-300',
  alternative: 'border border-gray-300 bg-white text-gray-900 hover:bg-gray-50 focus:ring-gray-200',
  light: 'border border-gray-200 bg-white text-gray-700 hover:bg-gray-50 focus:ring-gray-200',
  dark: 'bg-gray-950 text-white hover:bg-gray-800 focus:ring-gray-300',
  green: 'bg-green-700 text-white hover:bg-green-800 focus:ring-green-300',
  red: 'bg-red-700 text-white hover:bg-red-800 focus:ring-red-300',
};

const buttonSizeClasses: Record<ButtonSize, string> = {
  xs: 'h-8 px-3 text-xs',
  sm: 'h-9 px-3 text-sm',
  md: 'h-10 px-4 text-sm',
};

@Component({
  selector: 'app-button',
  imports: [RouterLink, Spinner],
  templateUrl: './button.component.html',
})
export class Button {
  color = input<ButtonColor>('primary');
  size = input<ButtonSize>('md');
  pill = input(false);
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

  protected className = computed(() =>
    classes(
      'inline-flex items-center justify-center gap-2 whitespace-nowrap font-medium focus:outline-none focus:ring-2 disabled:pointer-events-none disabled:opacity-60',
      buttonColorClasses[this.color()],
      buttonSizeClasses[this.size()],
      this.pill() ? 'rounded-full' : 'rounded-md',
      this.class(),
    ),
  );

  protected onLinkClick(event: MouseEvent) {
    if (this.disabled() || this.loading()) {
      event.preventDefault();
      return;
    }
    this.buttonClick.emit(event);
  }
}
