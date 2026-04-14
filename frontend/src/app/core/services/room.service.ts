import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { map, Observable } from 'rxjs';
import { API_URL } from '../config/endpoints';

export interface Room {
  id: string;
  code: string;
  name: string;
  gameType: string;
  smallBlind: number;
  bigBlind: number;
  maxPlayers: number;
  currentPlayers: number;
  isPrivate: boolean;
  isFull: boolean;
  seats: number;
}

export interface CreateRoomPayload {
  roomName: string;
  maxPlayers: number;
  smallBlind: number;
  bigBlind: number;
  isPrivate: boolean;
  roomPassword?: string;
}

const API = API_URL;

@Injectable({ providedIn: 'root' })
export class RoomService {
  constructor(private http: HttpClient) {}

  createRoom(payload: CreateRoomPayload): Observable<Room> {
    return this.http.post<Room>(`${API}/rooms`, payload, { withCredentials: true }).pipe(
      map((room) => this.withSeats(room))
    );
  }

  joinRoom(code: string, password?: string): Observable<Room> {
    const body = password ? { code, password } : { code };
    return this.http.post<Room>(`${API}/rooms/join`, body, { withCredentials: true }).pipe(
      map((room) => this.withSeats(room))
    );
  }

  getLiveRooms(): Observable<Room[]> {
    return this.http.get<Room[]>(`${API}/rooms?status=open`, { withCredentials: true }).pipe(
      map((rooms) => rooms.map((room) => this.withSeats(room)))
    );
  }

  private withSeats(room: Room): Room {
    return {
      ...room,
      seats: room.seats ?? Math.max(room.maxPlayers - room.currentPlayers, 0),
    };
  }
}
