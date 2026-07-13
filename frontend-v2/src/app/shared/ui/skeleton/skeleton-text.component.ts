import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';
import { Skeleton } from './skeleton.component';

@Component({
  selector: 'app-skeleton-text',
  imports: [Skeleton],
  templateUrl: './skeleton-text.component.html',
})
export class SkeletonText {
  lines = input(3);
  class = input<string>();

  protected lineIndexes = computed(() => Array.from({ length: this.lines() }, (_, index) => index));

  protected wrapperClass = computed(() => classes('space-y-3', this.class()));
}
