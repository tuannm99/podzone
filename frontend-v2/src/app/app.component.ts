import { Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { ToasterProvider } from './shared/ui/toaster-provider/toaster-provider.component';

@Component({
  selector: 'app-root',
  imports: [RouterOutlet, ToasterProvider],
  templateUrl: './app.component.html',
})
export class App {}
