import { Component, input, output } from '@angular/core';
import { MatTabsModule } from '@angular/material/tabs';

export type TabItem = {
  label: string;
  value: string;
};

// A tab *bar* only — it does not own or render tab content. Per
// agent/ANGULAR_STYLE_GUIDE.md's "Tab / section routing" rule, the active
// tab is state owned by the consumer (typically a route query param), not
// this component. Project the active panel's content via <ng-content/>;
// the consumer decides what that is based on `activeValue`.
@Component({
  selector: 'app-tabs',
  imports: [MatTabsModule],
  templateUrl: './tabs.component.html',
  styleUrl: './tabs.component.scss',
})
export class Tabs {
  tabs = input.required<TabItem[]>();
  activeValue = input.required<string>();

  activeValueChange = output<string>();
}
