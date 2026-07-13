import { Component, input } from '@angular/core';

@Component({
  selector: 'app-inline-message',
  templateUrl: './inline-message.component.html',
})
export class InlineMessage {
  when = input.required<boolean>();
  label = input.required<string>();
}
