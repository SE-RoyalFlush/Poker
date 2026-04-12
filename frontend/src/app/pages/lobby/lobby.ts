import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute } from '@angular/router';
import { Subject } from 'rxjs';
import { filter, takeUntil } from 'rxjs/operators';
import { uniqBy, orderBy } from 'lodash-es';

import { WebSocketService } from '../../core/services/websocket.service';
import { Player, WsMessage } from '../../core/models';

@Component({
  selector: 'app-lobby',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './lobby.html',
  styleUrl: './lobby.scss',
})
export class Lobby implements OnInit, OnDestroy {
  roomCode = '';
  players: Player[] = [];
  sortedPlayers: Player[] = [];

  private readonly destroy$ = new Subject<void>();

  constructor(
    private readonly route: ActivatedRoute,
    private readonly wsService: WebSocketService,
  ) {}

  ngOnInit(): void {
    this.roomCode = this.route.snapshot.queryParamMap.get('code') ?? '';

    this.wsService.connect();

    this.wsService.connected$
      .pipe(filter(v => v), takeUntil(this.destroy$))
      .subscribe(() => {
        if (this.roomCode) {
          this.wsService.sendMessage('JOIN_ROOM', { roomCode: this.roomCode });
        }
      });

    this.wsService.messages$
      .pipe(takeUntil(this.destroy$))
      .subscribe(msg => this.handleMessage(msg));
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
    this.wsService.disconnect();
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
    }

    if (playersChanged) {
      this.sortedPlayers = orderBy(this.players, ['isHost'], ['desc']);
    }
  }
}
