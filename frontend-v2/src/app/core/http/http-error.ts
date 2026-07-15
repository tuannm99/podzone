import { HttpErrorResponse } from '@angular/common/http';

type ServerErrorBody = { message?: string };

export function httpErrorMessage(error: unknown, fallback: string): string {
  if (error instanceof HttpErrorResponse) {
    const body = error.error as ServerErrorBody | null;
    return (
      (body && typeof body.message === 'string' ? body.message : undefined) ||
      error.message ||
      fallback
    );
  }
  if (error instanceof Error) {
    return error.message;
  }
  return fallback;
}
