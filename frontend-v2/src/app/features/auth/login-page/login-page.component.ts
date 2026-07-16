import { Component, inject, signal } from '@angular/core';
import { Router, RouterLink } from '@angular/router';

import { AuthService } from '../../../core/auth/auth.service';
import { Button } from '../../../shared/ui/button/button.component';
import { Card } from '../../../shared/ui/card/card.component';
import { ErrorAlert } from '../../../shared/ui/error-alert/error-alert.component';
import { InputField } from '../../../shared/ui/input-field/input-field.component';
import { SectionLead } from '../../../shared/ui/section-lead/section-lead.component';

@Component({
  selector: 'app-login-page',
  imports: [Card, InputField, Button, ErrorAlert, SectionLead, RouterLink],
  templateUrl: './login-page.component.html',
  styleUrl: './login-page.component.scss',
})
export class LoginPage {
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);

  username = signal('');
  password = signal('');
  error = signal('');
  submitting = signal(false);

  protected googleLoginUrl = this.auth.googleLoginUrl();

  async submit(event: SubmitEvent) {
    event.preventDefault();
    const username = this.username().trim();
    const password = this.password();
    if (!username || !password) return;

    this.submitting.set(true);
    this.error.set('');
    try {
      const result = await this.auth.login({ username, password });
      if (!result.success) {
        this.error.set(result.message);
        return;
      }
      await this.router.navigateByUrl('/');
    } finally {
      this.submitting.set(false);
    }
  }
}
