import { Component } from '@angular/core';
import { MatIconModule } from '@angular/material/icon';

@Component({
  selector: 'app-error-alert',
  imports: [MatIconModule],
  templateUrl: './error-alert.component.html',
  styleUrl: './error-alert.component.scss',
})
export class ErrorAlert {}
