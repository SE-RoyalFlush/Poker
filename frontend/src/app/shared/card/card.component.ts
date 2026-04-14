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
  private static readonly SUIT_SYMBOLS: Record<string, string> = {
    spades:   '♠',
    hearts:   '♥',
    diamonds: '♦',
    clubs:    '♣',
  };

  @Input() rank = '';
  @Input() suit = '';
  @Input() faceDown = false;

  get suitSymbol(): string {
    const lowerSuit = this.suit.toLowerCase();
    return CardComponent.SUIT_SYMBOLS[lowerSuit] ?? '';
  }

  get isRedSuit(): boolean {
    const lowerSuit = this.suit.toLowerCase();
    return lowerSuit === 'hearts' || lowerSuit === 'diamonds';
  }
}
