import { Component, input } from '@angular/core';

@Component({
  selector: 'app-code-block',
  templateUrl: './code-block.component.html',
})
export class CodeBlock {
  code = input.required<string>();
  label = input<string>();
  class = input<string>();
}
