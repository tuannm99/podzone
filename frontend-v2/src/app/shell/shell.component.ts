import { Component } from '@angular/core';
import { RouterLink, RouterOutlet } from '@angular/router';
import { NavLink } from '../shared/ui/nav-link/nav-link.component';

@Component({
  selector: 'app-shell',
  imports: [RouterOutlet, RouterLink, NavLink],
  templateUrl: './shell.component.html',
})
export class Shell {}
