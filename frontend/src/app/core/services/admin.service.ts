import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable, map } from 'rxjs';

export interface AdminUser {
  id: number;
  username: string;
}

@Injectable({ providedIn: 'root' })
export class AdminService {
  private readonly apiUrl = 'http://localhost:8080/api/admin';

  constructor(private http: HttpClient) {}

  private buildAuthHeader(username: string, password: string): HttpHeaders {
    const token = btoa(`${username}:${password}`);
    return new HttpHeaders({
      Authorization: `Basic ${token}`,
    });
  }

  getUsers(username: string, password: string): Observable<AdminUser[]> {
    return this.http
      .get<unknown>(`${this.apiUrl}/users`, {
        headers: this.buildAuthHeader(username, password),
        withCredentials: true,
      })
      .pipe(
        map((payload) => {
          const rawUsers = Array.isArray(payload)
            ? payload
            : Array.isArray((payload as { users?: unknown[] })?.users)
              ? (payload as { users: unknown[] }).users
              : [];

          return rawUsers
            .map((raw) => {
              const user = raw as {
                id?: number;
                ID?: number;
                username?: string;
                Username?: string;
              };

              const id = user.id ?? user.ID;
              const usernameValue = user.username ?? user.Username;

              if (typeof id !== 'number' || typeof usernameValue !== 'string') {
                return null;
              }

              return {
                id,
                username: usernameValue,
              } satisfies AdminUser;
            })
            .filter((user): user is AdminUser => user !== null);
        })
      );
  }

  deleteUser(adminUsername: string, adminPassword: string, userId: number): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/users/${userId}`, {
      headers: this.buildAuthHeader(adminUsername, adminPassword),
      withCredentials: true,
    });
  }
}
