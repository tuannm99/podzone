import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';

@Component({
  selector: 'app-app-shell',
  templateUrl: './app-shell.component.html',
})
export class AppShell {
  class = input<string>();
  containerClass = input<string>();

  protected outerClassName = computed(() =>
    classes('min-h-screen bg-gray-50 text-gray-900', this.class()),
  );
  protected innerClassName = computed(() => classes('pb-5 pt-0', this.containerClass()));
}
