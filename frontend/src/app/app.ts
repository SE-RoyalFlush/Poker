import { Component, OnInit } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { CsrfService } from './core/services/csrf.service';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [
    RouterOutlet
  ],
  templateUrl: './app.html',
  styleUrl: './app.scss'
})
export class App implements OnInit {
  constructor(private csrfService: CsrfService) {}

  ngOnInit(): void {
    this.csrfService.fetchToken().subscribe();
  }
}

