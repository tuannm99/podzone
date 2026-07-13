import { Component, computed, inject, input } from '@angular/core';
import { DomSanitizer } from '@angular/platform-browser';
import { classes } from '../../utils';

@Component({
  selector: 'app-video-embed',
  templateUrl: './video-embed.component.html',
})
export class VideoEmbed {
  private sanitizer = inject(DomSanitizer);

  title = input.required<string>();
  src = input.required<string>();
  aspect = input<'video' | 'wide' | 'square'>('video');
  class = input<string>();

  // iframe src is a resource-URL sanitization context in Angular —
  // must be explicitly trusted, unlike Solid which has no such gate.
  protected safeSrc = computed(() => this.sanitizer.bypassSecurityTrustResourceUrl(this.src()));

  protected wrapperClass = computed(() =>
    classes('overflow-hidden rounded-lg border border-gray-200 bg-black shadow-sm', this.class()),
  );

  protected aspectClass = computed(() => {
    if (this.aspect() === 'square') return 'aspect-square';
    if (this.aspect() === 'wide') return 'aspect-[21/9]';
    return 'aspect-video';
  });

  protected iframeClass = computed(() => classes('w-full', this.aspectClass()));
}
