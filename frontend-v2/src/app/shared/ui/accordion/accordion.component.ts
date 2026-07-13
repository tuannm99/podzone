import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';

// NOTE (port judgment call): Solid's AccordionItem.content is `JSX.Element`
// — arbitrary rich content per item. Angular has no direct equivalent for
// "rich content inside a data array" without a TemplateRef-based API, which
// is a bigger structural change than a faithful port. Typed as `string`
// here; upgrade to a `TemplateRef`-per-item API if rich per-item content is
// actually needed later.
export type AccordionItem = {
  title: string;
  content: string;
  description?: string;
  defaultOpen?: boolean;
  badge?: string;
};

@Component({
  selector: 'app-accordion',
  templateUrl: './accordion.component.html',
})
export class Accordion {
  items = input.required<AccordionItem[]>();
  class = input<string>();

  protected wrapperClass = computed(() =>
    classes(
      'divide-y divide-gray-200 overflow-hidden rounded-lg border border-gray-200 bg-white',
      this.class(),
    ),
  );
}
