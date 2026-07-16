import { Component, computed, inject, signal } from '@angular/core';
import { Router, RouterLink } from '@angular/router';

import { AuthService } from '../../../core/auth/auth.service';
import { Button } from '../../../shared/ui/button/button.component';
import { Card } from '../../../shared/ui/card/card.component';
import { ErrorAlert } from '../../../shared/ui/error-alert/error-alert.component';
import { InputField } from '../../../shared/ui/input-field/input-field.component';
import { SectionLead } from '../../../shared/ui/section-lead/section-lead.component';

@Component({
  selector: 'app-register-page',
  imports: [Card, InputField, Button, ErrorAlert, SectionLead, RouterLink],
  templateUrl: './register-page.component.html',
  styleUrl: './register-page.component.scss',
})
export class RegisterPage {
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);

  username = signal('');
  email = signal('');
  password = signal('');
  confirmPassword = signal('');
  error = signal('');
  submitting = signal(false);

  protected confirmPasswordMismatch = computed(
    () => this.confirmPassword().length > 0 && this.confirmPassword() !== this.password(),
  );

  protected canSubmit = computed(
    () =>
      !!this.username().trim() &&
      !!this.email().trim() &&
      !!this.password() &&
      !!this.confirmPassword() &&
      !this.confirmPasswordMismatch(),
  );

  async submit(event: SubmitEvent) {
    event.preventDefault();
    if (!this.canSubmit()) return;

    this.submitting.set(true);
    this.error.set('');
    try {
      const result = await this.auth.register({
        username: this.username().trim(),
        email: this.email().trim(),
        password: this.password(),
      });
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
