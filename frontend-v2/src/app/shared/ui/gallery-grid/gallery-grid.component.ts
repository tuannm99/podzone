import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';

export type GalleryItem = {
  src: string;
  alt: string;
  caption?: string;
};

const columnClasses = {
  2: 'sm:grid-cols-2',
  3: 'sm:grid-cols-2 lg:grid-cols-3',
  4: 'sm:grid-cols-2 lg:grid-cols-4',
} as const;

@Component({
  selector: 'app-gallery-grid',
  templateUrl: './gallery-grid.component.html',
})
export class GalleryGrid {
  items = input.required<GalleryItem[]>();
  columns = input<2 | 3 | 4>(3);
  class = input<string>();

  protected className = computed(() =>
    classes('grid gap-4', columnClasses[this.columns()], this.class()),
  );
}
