import { Component, input } from '@angular/core';
import { Spinner } from '../spinner/spinner.component';

@Component({
  selector: 'app-loading-inline',
  imports: [Spinner],
  templateUrl: './loading-inline.component.html',
  styleUrl: './loading-inline.component.scss',
})
export class LoadingInline {
  label = input.required<string>();
}
