import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { API_URL } from '../config/endpoints';

export interface UserStats {
  handsPlayed: number;
  wins: number;
  losses: number;
  winRate: number;
  totalEarnings: number;
}

@Injectable({ providedIn: 'root' })
export class StatsService {
  constructor(private http: HttpClient) {}

  getUserStats(userId: string): Observable<UserStats> {
    return this.http.get<UserStats>(`${API_URL}/users/${userId}/stats`, {
      withCredentials: true,
    });
  }
}
