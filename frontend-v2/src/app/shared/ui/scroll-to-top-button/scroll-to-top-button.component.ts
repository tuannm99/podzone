import { Component, computed, effect, input, signal } from '@angular/core';
import { classes } from '../../utils';
import { Button } from '../button/button.component';

@Component({
  selector: 'app-scroll-to-top-button',
  imports: [Button],
  templateUrl: './scroll-to-top-button.component.html',
})
export class ScrollToTopButton {
  threshold = input(280);
  class = input<string>();

  protected visible = signal(false);

  protected className = computed(() =>
    classes(
      'fixed bottom-6 right-6 z-30 shadow-lg transition duration-200',
      this.visible() ? 'translate-y-0 opacity-100' : 'pointer-events-none translate-y-3 opacity-0',
      this.class(),
    ),
  );

  constructor() {
    // effect() here is external DOM synchronization (a scroll listener),
    // the sanctioned use per ANGULAR_STYLE_GUIDE.md — not a data fetch.
    effect((onCleanup) => {
      const threshold = this.threshold();
      const updateVisibility = () => this.visible.set(window.scrollY > threshold);
      updateVisibility();
      window.addEventListener('scroll', updateVisibility, { passive: true });
      onCleanup(() => window.removeEventListener('scroll', updateVisibility));
    });
  }

  protected scrollToTop() {
    window.scrollTo({ top: 0, left: 0, behavior: 'smooth' });
  }
}
