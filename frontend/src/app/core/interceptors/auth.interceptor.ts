import { Injectable } from '@angular/core';
import {
  HttpInterceptor,
  HttpRequest,
  HttpHandler,
  HttpEvent,
  HttpErrorResponse
} from '@angular/common/http';
import { Observable, throwError } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { Router } from '@angular/router';
import { CsrfService } from '../services/csrf.service';

@Injectable()
export class AuthInterceptor implements HttpInterceptor {

  constructor(
    private router: Router,
    private csrfService: CsrfService
  ) {}

  intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    const mutatingMethods = ['POST', 'PUT', 'PATCH', 'DELETE'];
    const csrfToken = this.csrfService.getToken();
    const isSessionCheckRequest = /\/api\/me(?:\?|$)/.test(req.url);

    const modifiedReq = req.clone({
      withCredentials: true,
      ...(mutatingMethods.includes(req.method) && csrfToken
        ? { headers: req.headers.set('X-CSRF-Token', csrfToken) }
        : {})
    });

    return next.handle(modifiedReq).pipe(
      catchError((error: HttpErrorResponse) => {
        if (error.status === 401) {
          this.csrfService.clearToken();
          if (!isSessionCheckRequest) {
            this.router.navigate(['/login']);
          }
        }
        return throwError(() => error);
      })
    );
  }
}
