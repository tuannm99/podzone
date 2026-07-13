import { Component, TemplateRef, computed, effect, input, signal } from '@angular/core';
import { NgTemplateOutlet } from '@angular/common';
import { classes } from '../../utils';

export type CarouselSlide = {
  id?: string;
  eyebrow?: string;
  title?: string;
  copy?: string;
  imageSrc?: string;
  imageAlt?: string;
  content?: TemplateRef<unknown>;
  action?: TemplateRef<unknown>;
};

@Component({
  selector: 'app-carousel',
  imports: [NgTemplateOutlet],
  templateUrl: './carousel.component.html',
})
export class Carousel {
  slides = input.required<CarouselSlide[]>();
  autoPlay = input(false);
  intervalMs = input(5000);
  class = input<string>();

  protected currentIndex = signal(0);

  protected className = computed(() =>
    classes('overflow-hidden rounded-lg border border-gray-200 bg-white shadow-sm', this.class()),
  );

  constructor() {
    effect((onCleanup) => {
      if (!this.autoPlay() || this.slides().length <= 1) return;

      const timer = window.setInterval(() => {
        this.currentIndex.update((index) => (index + 1) % this.slides().length);
      }, this.intervalMs());

      onCleanup(() => window.clearInterval(timer));
    });
  }

  protected goTo(index: number) {
    this.currentIndex.set(index);
  }

  protected previous() {
    this.currentIndex.update((index) => (index - 1 + this.slides().length) % this.slides().length);
  }

  protected next() {
    this.currentIndex.update((index) => (index + 1) % this.slides().length);
  }
}
