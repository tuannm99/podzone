import { Component, computed, input } from '@angular/core';
import { classes } from '../../utils';

@Component({
  selector: 'app-data-table',
  templateUrl: './data-table.component.html',
})
export class DataTable {
  class = input<string>();

  protected className = computed(() =>
    classes('overflow-x-auto rounded-lg border border-gray-200', this.class()),
  );
}
