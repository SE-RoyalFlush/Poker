import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of, tap, map } from 'rxjs';

interface CsrfResponse {
  csrfToken: string;
}

@Injectable({ providedIn: 'root' })
export class CsrfService {

  private token: string | null = null;
  private readonly csrfUrl = 'http://localhost:8080/api/csrf';

  constructor(private http: HttpClient) {}

  fetchToken(): Observable<CsrfResponse> {
    return this.http.get<CsrfResponse>(this.csrfUrl, {
      withCredentials: true
    }).pipe(
      tap(response => {
        this.token = response.csrfToken;
      })
    );
  }

  getToken(): string | null {
    return this.token;
  }

  ensureToken(): Observable<string> {
    if (this.token) {
      return of(this.token);
    }

    return this.fetchToken().pipe(
      map(response => response.csrfToken)
    );
  }

  clearToken(): void {
    this.token = null;
  }
}
