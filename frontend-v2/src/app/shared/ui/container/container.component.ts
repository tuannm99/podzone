import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';

export type ContainerWidth = 'lg' | 'xl' | '2xl' | '7xl';

const widthClasses: Record<ContainerWidth, string> = {
  lg: 'max-w-5xl',
  xl: 'max-w-6xl',
  '2xl': 'max-w-[96rem]',
  '7xl': 'max-w-7xl',
};

@Component({
  selector: 'app-container',
  templateUrl: './container.component.html',
})
export class Container {
  class = input<string>();
  width = input<ContainerWidth>('2xl');

  protected className = computed(() =>
    classes('mx-auto w-full px-4 sm:px-6 lg:px-8', widthClasses[this.width()], this.class()),
  );
}
