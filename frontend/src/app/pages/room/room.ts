import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute } from '@angular/router';

@Component({
  selector: 'app-room',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './room.html',
  styleUrl: './room.scss',
})
export class RoomComponent implements OnInit {
  roomCode = '';

  constructor(private route: ActivatedRoute) {}

  ngOnInit(): void {
    this.roomCode = this.route.snapshot.paramMap.get('code') ?? '';
  }
}
