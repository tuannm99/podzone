import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';

@Component({
  selector: 'app-device-mockup',
  templateUrl: './device-mockup.component.html',
})
export class DeviceMockup {
  label = input<string>();
  class = input<string>();
  screenClass = input<string>();

  protected className = computed(() =>
    classes('inline-flex flex-col items-center gap-3', this.class()),
  );
  protected screenClasses = computed(() =>
    classes('min-h-96 w-[20rem] overflow-hidden rounded-[2rem] bg-white', this.screenClass()),
  );
}
