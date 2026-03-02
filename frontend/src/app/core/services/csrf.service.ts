import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, tap } from 'rxjs';

interface CsrfResponse {
  csrfToken: string;
}

@Injectable({ providedIn: 'root' })
export class CsrfService {

  private token: string | null = null;
  private readonly csrfUrl = 'http://localhost:8080/api/csrf';

  constructor(private http: HttpClient) {}

  fetchToken(): Observable<CsrfResponse> {
    console.log('fetchToken called, hitting:', this.csrfUrl);
    return this.http.get<CsrfResponse>(this.csrfUrl, {
      withCredentials: true
    }).pipe(
      tap(response => {
        console.log('token stored:', response.csrfToken);
        this.token = response.csrfToken;
      })
    );
  }

  getToken(): string | null {
    return this.token;
  }

  clearToken(): void {
    this.token = null;
  }
}
