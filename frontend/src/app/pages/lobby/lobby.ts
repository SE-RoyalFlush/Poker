import { Component, OnInit, OnDestroy, ViewChild, ElementRef, AfterViewChecked } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { Subject } from 'rxjs';
import { filter, takeUntil } from 'rxjs/operators';
import { uniqBy, orderBy } from 'lodash-es';

import { MatListModule } from '@angular/material/list';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';

import { WebSocketService } from '../../core/services/websocket.service';
import { AuthService } from '../../core/services/auth.service';
import { Player, ChatMessage, WsMessage } from '../../core/models';

@Component({
  selector: 'app-lobby',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    MatListModule,
    MatInputModule,
    MatButtonModule,
    MatFormFieldModule,
  ],
  templateUrl: './lobby.html',
  styleUrl: './lobby.scss',
})
export class Lobby implements OnInit, OnDestroy, AfterViewChecked {
  @ViewChild('chatScroll') private chatScrollRef!: ElementRef<HTMLElement>;

  roomCode = '';
  players: Player[] = [];
  sortedPlayers: Player[] = [];
  chatMessages: ChatMessage[] = [];
  chatInput = '';
  isReady = false;

  private currentUsername = '';
  private shouldScrollChat = false;
  private readonly destroy$ = new Subject<void>();

  constructor(
    private readonly route: ActivatedRoute,
    private readonly router: Router,
    private readonly wsService: WebSocketService,
    private readonly authService: AuthService,
  ) {}

  ngOnInit(): void {
    this.roomCode = this.route.snapshot.queryParamMap.get('code') ?? '';

    if (!this.roomCode) {
      this.router.navigate(['/dashboard']);
      return;
    }

    this.authService.currentUser$
      .pipe(takeUntil(this.destroy$))
      .subscribe(user => {
        this.currentUsername = user?.username ?? '';
      });

    this.wsService.connect();

    this.wsService.connected$
      .pipe(filter(v => v), takeUntil(this.destroy$))
      .subscribe(() => {
        this.wsService.sendMessage('JOIN_ROOM', { roomCode: this.roomCode });
      });

    this.wsService.messages$
      .pipe(takeUntil(this.destroy$))
      .subscribe(msg => this.handleMessage(msg));
  }

  ngAfterViewChecked(): void {
    if (this.shouldScrollChat) {
      this.scrollChatToBottom();
      this.shouldScrollChat = false;
    }
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
    this.wsService.disconnect();
  }

  toggleReady(): void {
    this.isReady = !this.isReady;
    this.wsService.sendMessage('TOGGLE_READY', { isReady: this.isReady });
  }

  sendChat(): void {
    const text = this.chatInput.trim();
    if (!text) return;
    this.wsService.sendMessage('CHAT_MESSAGE', { text });
    this.chatInput = '';
  }

  onChatKeydown(event: KeyboardEvent): void {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault();
      this.sendChat();
    }
  }

  private handleMessage(msg: WsMessage): void {
    let playersChanged = false;

    if (msg.type === 'PLAYER_JOINED') {
      const player = msg.payload as Player;
      this.players = uniqBy([...this.players, player], 'id');
      playersChanged = true;
    } else if (msg.type === 'PLAYER_LEFT') {
      const { id } = msg.payload as Pick<Player, 'id'>;
      this.players = this.players.filter(p => p.id !== id);
      playersChanged = true;
    } else if (msg.type === 'PLAYER_READY') {
      const { id, isReady } = msg.payload as Pick<Player, 'id' | 'isReady'>;
      this.players = this.players.map(p => p.id === id ? { ...p, isReady } : p);
      playersChanged = true;
    } else if (msg.type === 'CHAT_MESSAGE') {
      const chatMsg = msg.payload as ChatMessage;
      this.chatMessages = [...this.chatMessages, chatMsg];
      this.shouldScrollChat = true;
    }

    if (playersChanged) {
      this.sortedPlayers = orderBy(this.players, ['isHost'], ['desc']);
    }
  }

  private scrollChatToBottom(): void {
    if (this.chatScrollRef?.nativeElement) {
      this.chatScrollRef.nativeElement.scrollTop =
        this.chatScrollRef.nativeElement.scrollHeight;
    }
  }
}
