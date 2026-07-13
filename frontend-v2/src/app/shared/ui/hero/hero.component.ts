import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';

// NOTE (port judgment call): Solid's `export const Jumbotron = Hero` alias
// is not ported — Angular components are classes bound to one selector, so
// a second identical component with a different selector would just be a
// duplicate file. Use <app-hero> for both cases.
@Component({
  selector: 'app-hero',
  templateUrl: './hero.component.html',
})
export class Hero {
  eyebrow = input<string>();
  title = input.required<string>();
  copy = input<string>();
  class = input<string>();

  protected wrapperClass = computed(() =>
    classes('rounded-lg border border-gray-200 bg-white p-8 shadow-sm sm:p-10', this.class()),
  );
}
