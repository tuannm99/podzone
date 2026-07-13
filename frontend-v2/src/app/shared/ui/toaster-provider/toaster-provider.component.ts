import { Component, inject } from '@angular/core';
import { ToasterService, type ToastType } from '../../services/toaster.service';
import { classes } from '../../utils';

const toastColorClasses: Record<ToastType, string> = {
  success: 'bg-green-700 text-white',
  error: 'bg-red-700 text-white',
  info: 'bg-gray-800 text-white',
  warning: 'bg-yellow-600 text-white',
};

@Component({
  selector: 'app-toaster-provider',
  templateUrl: './toaster-provider.component.html',
})
export class ToasterProvider {
  protected toaster = inject(ToasterService);

  protected toastClass(type: ToastType) {
    return classes(
      'flex items-center gap-3 rounded-lg px-4 py-3 shadow-lg text-sm font-medium',
      toastColorClasses[type],
    );
  }
}
