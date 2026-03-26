import { Injectable } from '@angular/core';
import {
  HttpInterceptor,
  HttpRequest,
  HttpHandler,
  HttpEvent,
  HttpErrorResponse
} from '@angular/common/http';
import { Observable, throwError } from 'rxjs';
import { catchError, switchMap } from 'rxjs/operators';
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
    const isMutatingRequest = mutatingMethods.includes(req.method);
    const csrfToken = this.csrfService.getToken();
    const isSessionCheckRequest = /\/api\/me(?:\?|$)/.test(req.url);

    const createRequest = (token?: string | null): HttpRequest<any> => {
      let cloned = req.clone({ withCredentials: true });
      if (isMutatingRequest && token) {
        cloned = cloned.clone({ headers: cloned.headers.set('X-CSRF-Token', token) });
      }
      return cloned;
    };

    const handle = (r: HttpRequest<any>) => next.handle(r).pipe(
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

    if (isMutatingRequest && !csrfToken) {
      return this.csrfService.ensureToken().pipe(
        switchMap(token => handle(createRequest(token)))
      );
    }

    return handle(createRequest(csrfToken));
  }
}
