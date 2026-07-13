import { Injectable, signal } from '@angular/core';

// Ported from frontend/packages/shared/ui/toaster/createToaster.ts — Solid's
// module-level signal singleton becomes an Angular `providedIn: 'root'`
// service; same push/dismiss/duration behavior.
export type ToastType = 'success' | 'error' | 'info' | 'warning';

export type Toast = {
  id: string;
  type: ToastType;
  message: string;
};

@Injectable({ providedIn: 'root' })
export class ToasterService {
  private counter = 0;
  private readonly _toasts = signal<Toast[]>([]);
  readonly toasts = this._toasts.asReadonly();

  success(message: string, durationMs?: number) {
    this.push('success', message, durationMs);
  }

  error(message: string, durationMs?: number) {
    this.push('error', message, durationMs);
  }

  info(message: string, durationMs?: number) {
    this.push('info', message, durationMs);
  }

  warning(message: string, durationMs?: number) {
    this.push('warning', message, durationMs);
  }

  dismiss(id: string) {
    this._toasts.update((prev) => prev.filter((toast) => toast.id !== id));
  }

  private push(type: ToastType, message: string, durationMs = 4000) {
    const id = `toast-${++this.counter}`;
    this._toasts.update((prev) => [...prev, { id, type, message }]);
    setTimeout(() => this.dismiss(id), durationMs);
  }
}
