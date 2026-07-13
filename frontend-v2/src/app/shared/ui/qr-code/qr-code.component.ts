import { Component, computed, inject, input, resource } from '@angular/core';
import { DomSanitizer } from '@angular/platform-browser';
import QRCode from 'qrcode';
import { classes } from '../../utils';

@Component({
  selector: 'app-qr-code',
  templateUrl: './qr-code.component.html',
})
export class QrCode {
  value = input.required<string>();
  size = input<number>();
  class = input<string>();
  panelClass = input<string>();

  private sanitizer = inject(DomSanitizer);

  protected className = computed(() =>
    classes('inline-flex rounded-lg border border-gray-200 bg-white p-4 shadow-sm', this.class()),
  );
  protected emptyClass = computed(() =>
    classes(
      'flex min-h-48 min-w-48 items-center justify-center rounded-lg bg-gray-50 px-4 py-4 text-center text-sm text-gray-500',
      this.panelClass(),
    ),
  );

  protected qr = resource({
    params: () => {
      const trimmed = this.value().trim();
      return trimmed ? { value: trimmed, size: this.size() ?? 192 } : undefined;
    },
    loader: ({ params }) =>
      QRCode.toString(params.value, {
        type: 'svg',
        width: params.size,
        margin: 1,
        errorCorrectionLevel: 'M',
        color: {
          dark: '#111827',
          light: '#FFFFFFFF',
        },
      }),
  });

  protected safeSvg = computed(() => {
    const svg = this.qr.value();
    return svg ? this.sanitizer.bypassSecurityTrustHtml(svg) : null;
  });

  protected errorMessage = computed(() => {
    const error = this.qr.error();
    if (!error) return '';
    return error instanceof Error ? error.message : 'Unable to generate QR code';
  });
}
