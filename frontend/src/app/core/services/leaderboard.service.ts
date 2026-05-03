import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { API_URL } from '../config/endpoints';

export interface LeaderboardEntry {
  rank: number;
  userId: number;
  username: string;
  handsPlayed: number;
  wins: number;
  winRate: number;
  totalEarnings: number;
}

@Injectable({ providedIn: 'root' })
export class LeaderboardService {
  constructor(private http: HttpClient) {}

  getLeaderboard(): Observable<LeaderboardEntry[]> {
    return this.http.get<LeaderboardEntry[]>(`${API_URL}/leaderboard`, {
      withCredentials: true,
    });
  }
}
