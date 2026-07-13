import { Component, computed, input, signal } from '@angular/core';
import { classes } from '../../utils';

@Component({
  selector: 'app-clipboard-button',
  templateUrl: './clipboard-button.component.html',
})
export class ClipboardButton {
  text = input.required<string>();
  label = input<string>();
  copiedLabel = input<string>();
  class = input<string>();

  protected copied = signal(false);

  protected className = computed(() =>
    classes(
      'inline-flex items-center gap-2 rounded-md border border-gray-200 bg-white px-4 py-2.5 text-sm font-medium text-gray-700 shadow-sm transition hover:bg-gray-50',
      this.copied() && 'border-green-200 bg-green-50 text-green-700',
      this.class(),
    ),
  );

  protected displayLabel = computed(() =>
    this.copied() ? (this.copiedLabel() ?? 'Copied') : (this.label() ?? 'Copy'),
  );

  protected async handleCopy() {
    try {
      await navigator.clipboard.writeText(this.text());
      this.copied.set(true);
      window.setTimeout(() => this.copied.set(false), 1500);
    } catch {
      this.copied.set(false);
    }
  }
}
