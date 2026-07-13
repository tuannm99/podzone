import { Component, input } from '@angular/core';
import { LoadingInline } from '../loading-inline/loading-inline.component';

@Component({
  selector: 'app-loading-block',
  imports: [LoadingInline],
  templateUrl: './loading-block.component.html',
})
export class LoadingBlock {
  label = input.required<string>();
}
