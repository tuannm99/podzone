import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';

export type HeadingTag = 'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'h6';

const headingClasses: Record<HeadingTag, string> = {
  h1: 'text-4xl font-semibold tracking-tight sm:text-5xl',
  h2: 'text-3xl font-semibold tracking-tight sm:text-4xl',
  h3: 'text-2xl font-semibold tracking-tight',
  h4: 'text-xl font-semibold tracking-tight',
  h5: 'text-lg font-semibold',
  h6: 'text-base font-semibold uppercase tracking-wide',
};

@Component({
  selector: 'app-heading',
  templateUrl: './heading.component.html',
})
export class Heading {
  // Named `tag`, not `as` — `as` is a reserved word in Angular template
  // expression syntax (used for `@if (...; as x)` aliasing) and cannot be
  // used as a method/property name in a template expression like `@switch`.
  tag = input<HeadingTag>('h2');
  class = input<string>();

  protected className = computed(() =>
    classes('text-gray-900', headingClasses[this.tag()], this.class()),
  );
}
