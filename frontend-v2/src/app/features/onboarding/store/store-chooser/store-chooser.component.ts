import { Component, input, output } from '@angular/core';

import { Badge } from '../../../../shared/ui/badge/badge.component';
import { Button } from '../../../../shared/ui/button/button.component';
import { EmptyBlock } from '../../../../shared/ui/empty-block/empty-block.component';
import { ErrorAlert } from '../../../../shared/ui/error-alert/error-alert.component';
import { InputField } from '../../../../shared/ui/input-field/input-field.component';
import { LoadingBlock } from '../../../../shared/ui/loading-block/loading-block.component';
import {
  isReadyToOpen,
  isRetryableStatus,
  readinessTone,
  storeStatusLabel,
  type StoreReadiness,
  type StoreRequest,
} from '../store.types';

@Component({
  selector: 'app-store-chooser',
  imports: [LoadingBlock, ErrorAlert, EmptyBlock, Badge, Button, InputField],
  templateUrl: './store-chooser.component.html',
  styleUrl: './store-chooser.component.scss',
})
export class StoreChooser {
  tenantId = input.required<string>();
  stores = input.required<StoreRequest[]>();
  readiness = input<Record<string, StoreReadiness | undefined>>({});
  loading = input(false);
  error = input('');
  retryingId = input('');
  creating = input(false);
  newStoreName = input('');

  newStoreNameChange = output<string>();
  createStore = output<void>();
  retryStore = output<string>();
  openStore = output<StoreRequest>();

  protected readonly label = storeStatusLabel;
  protected readonly retryable = isRetryableStatus;

  protected tone(store: StoreRequest) {
    return readinessTone(store.status, this.readiness()[store.id]?.ui_state);
  }

  protected openable(store: StoreRequest): boolean {
    return isReadyToOpen(store, this.readiness()[store.id]?.ui_state);
  }

  protected onCreateSubmit(event: SubmitEvent) {
    event.preventDefault();
    if (!this.newStoreName().trim() || this.creating()) return;
    this.createStore.emit();
  }
}
