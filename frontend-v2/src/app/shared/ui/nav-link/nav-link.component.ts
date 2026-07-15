import { Component, input } from '@angular/core';
import { RouterLink, RouterLinkActive } from '@angular/router';

@Component({
  selector: 'app-nav-link',
  imports: [RouterLink, RouterLinkActive],
  templateUrl: './nav-link.component.html',
  styleUrl: './nav-link.component.scss',
})
export class NavLink {
  href = input.required<string>();
  exact = input(false);
  tag = input<string>();
}
