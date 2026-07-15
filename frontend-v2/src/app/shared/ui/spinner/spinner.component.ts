import { Component, input } from '@angular/core';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';

@Component({
  selector: 'app-spinner',
  imports: [MatProgressSpinnerModule],
  template: `<mat-progress-spinner mode="indeterminate" [diameter]="diameter()" />`,
})
export class Spinner {
  diameter = input(16);
}
