import { Component, inject } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { classes } from '../../utils';
import { ToasterService, type ToastType } from '../../services/toaster.service';

const toastToneClasses: Record<ToastType, string> = {
  success: 'toast--success',
  error: 'toast--error',
  info: 'toast--info',
  warning: 'toast--warning',
};

const toastIcons: Record<ToastType, string> = {
  success: 'check_circle',
  error: 'error',
  info: 'info',
  warning: 'warning',
};

@Component({
  selector: 'app-toaster-provider',
  imports: [MatIconModule, MatButtonModule],
  templateUrl: './toaster-provider.component.html',
  styleUrl: './toaster-provider.component.scss',
})
export class ToasterProvider {
  protected toaster = inject(ToasterService);

  protected toastClass(type: ToastType) {
    return classes('toast', toastToneClasses[type]);
  }

  protected toastIcon(type: ToastType) {
    return toastIcons[type];
  }
}
