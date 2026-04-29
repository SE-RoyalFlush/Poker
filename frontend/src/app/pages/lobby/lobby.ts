import { Component, OnInit, OnDestroy, ViewChild, ElementRef, AfterViewChecked } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { Subject } from 'rxjs';
import { filter, takeUntil } from 'rxjs/operators';
import { uniqBy, orderBy } from 'lodash-es';

import { MatListModule } from '@angular/material/list';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';

import { AuthService } from '../../core/services/auth.service';
import { WebSocketService } from '../../core/services/websocket.service';
import {
  ClientMessageType,
  Player,
  PlayerLeftPayload,
  RoomStatePayload,
  ServerMessageType,
  WsMessage,
} from '../../core/models';

export interface ChatMessage {
  sender: string;
  text: string;
  timestamp: string;
}

@Component({
  selector: 'app-lobby',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    MatListModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatIconModule,
  ],
  templateUrl: './lobby.html',
  styleUrl: './lobby.scss',
})
export class Lobby implements OnInit, OnDestroy, AfterViewChecked {
  @ViewChild('chatScroll') chatScrollEl?: ElementRef<HTMLDivElement>;

  roomCode = '';
  players: Player[] = [];
  sortedPlayers: Player[] = [];

  isReady = false;
  messages: ChatMessage[] = [];
  chatInput = '';
  currentUsername = '';

  private shouldScroll = false;
  private readonly destroy$ = new Subject<void>();

  constructor(
    private readonly route: ActivatedRoute,
    private readonly router: Router,
    private readonly authService: AuthService,
    private readonly wsService: WebSocketService,
  ) {}

  ngOnInit(): void {
    this.roomCode = this.route.snapshot.paramMap.get('code') ?? '';

    if (!this.roomCode) {
      this.router.navigate(['/dashboard']);
      return;
    }

    const user = this.authService.getCurrentUser();
    this.currentUsername = user?.username ?? 'You';

    this.wsService.connect();

    this.wsService.connected$
      .pipe(filter(v => v), takeUntil(this.destroy$))
      .subscribe(() => {
        this.wsService.sendMessage(ClientMessageType.JOIN_ROOM, { roomCode: this.roomCode });
      });

    this.wsService.messages$
      .pipe(takeUntil(this.destroy$))
      .subscribe(msg => this.handleMessage(msg));
  }

  ngAfterViewChecked(): void {
    if (this.shouldScroll) {
      this.scrollChatToBottom();
      this.shouldScroll = false;
    }
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
    this.wsService.disconnect();
  }

  private handleMessage(msg: WsMessage): void {
    let playersChanged = false;

    if (msg.type === ServerMessageType.ROOM_STATE) {
      const payload = msg.payload as RoomStatePayload;
      this.players = uniqBy(payload.players, 'id');
      playersChanged = true;
    } else if (msg.type === ServerMessageType.PLAYER_JOINED) {
      const player = msg.payload as Player;
      this.players = uniqBy([...this.players, player], 'id');
      playersChanged = true;
    } else if (msg.type === ServerMessageType.PLAYER_LEFT) {
      const { id } = msg.payload as PlayerLeftPayload;
      this.players = this.players.filter(p => p.id !== id);
      playersChanged = true;
    } else if (msg.type === ServerMessageType.PLAYER_UPDATE) {
      const player = msg.payload as Player;
      this.players = this.players.map(p => p.id === player.id ? { ...p, ...player } : p);
      playersChanged = true;
    }

    if (playersChanged) {
      this.sortedPlayers = orderBy(this.players, ['isHost'], ['desc']);
    }
  }

  toggleReady(): void {
    this.isReady = !this.isReady;
    this.wsService.sendMessage(ClientMessageType.TOGGLE_READY, {});
  }

  sendMessage(): void {
    const text = this.chatInput.trim();
    if (!text) return;

    const msg: ChatMessage = {
      sender: this.currentUsername,
      text,
      timestamp: new Date().toISOString(),
    };
    this.messages.push(msg);
    this.wsService.sendMessage('CHAT_MESSAGE', { text });
    this.chatInput = '';
    this.shouldScroll = true;
  }

  onChatKeydown(event: KeyboardEvent): void {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault();
      this.sendMessage();
    }
  }

  formatTime(timestamp: string): string {
    const d = new Date(timestamp);
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  }

  private scrollChatToBottom(): void {
    const el = this.chatScrollEl?.nativeElement;
    if (el) el.scrollTop = el.scrollHeight;
  }
}
