import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';
import { avatarSizeClasses, initials, type AvatarSize } from '../display-shared';
import { Indicator, type IndicatorColor } from '../indicator/indicator.component';

export type { AvatarSize };

@Component({
  selector: 'app-avatar',
  imports: [Indicator],
  templateUrl: './avatar.component.html',
})
export class Avatar {
  src = input<string>();
  alt = input<string>();
  name = input<string>();
  size = input<AvatarSize>('md');
  rounded = input(true);
  status = input<'online' | 'offline' | 'busy' | 'away'>();
  class = input<string>();

  protected initialsText = computed(() => initials(this.name()));

  protected statusColor = computed<IndicatorColor>(() => {
    const status = this.status();
    if (status === 'busy') return 'red';
    if (status === 'away') return 'yellow';
    if (status === 'online') return 'green';
    return 'gray';
  });

  protected wrapperClass = computed(() => classes('relative inline-flex', this.class()));

  protected innerClass = computed(() =>
    classes(
      'inline-flex items-center justify-center overflow-hidden bg-gray-200 font-semibold text-gray-700',
      avatarSizeClasses[this.size()],
      this.rounded() ? 'rounded-full' : 'rounded-lg',
    ),
  );
}
