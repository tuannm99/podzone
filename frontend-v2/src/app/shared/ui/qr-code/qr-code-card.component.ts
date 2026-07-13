import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';
import { QrCode } from './qr-code.component';

@Component({
  selector: 'app-qr-code-card',
  imports: [QrCode],
  templateUrl: './qr-code-card.component.html',
})
export class QrCodeCard {
  value = input.required<string>();
  title = input<string>();
  copy = input<string>();
  class = input<string>();

  protected className = computed(() =>
    classes(
      'inline-flex flex-col gap-4 rounded-lg border border-gray-200 bg-white p-5 shadow-sm',
      this.class(),
    ),
  );
}
