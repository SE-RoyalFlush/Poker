import { Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-card',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './card.component.html',
  styleUrl: './card.component.scss',
})
export class CardComponent {
  @Input() rank = '';
  @Input() suit = '';
  @Input() faceDown = false;

  get suitSymbol(): string {
    const symbols: Record<string, string> = {
      spades:   '♠',
      hearts:   '♥',
      diamonds: '♦',
      clubs:    '♣',
    };
    return symbols[this.suit.toLowerCase()] ?? '';
  }

  get isRedSuit(): boolean {
    const s = this.suit.toLowerCase();
    return s === 'hearts' || s === 'diamonds';
  }
}
