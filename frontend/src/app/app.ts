import { Component, OnInit } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatButtonModule } from '@angular/material/button';
import { CsrfService } from './core/services/csrf.service';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [
    RouterOutlet,
    MatToolbarModule,
    MatButtonModule
  ],
  templateUrl: './app.html',
  styleUrl: './app.scss'
})
export class App implements OnInit {
  title = 'RoyalFlush';

  constructor(private csrfService: CsrfService) {}

  ngOnInit(): void {
    console.log('ngOnInit fired');
    console.log('csrfService:', this.csrfService);

    this.csrfService.fetchToken().subscribe({
      next: (data) => console.log('CSRF token fetched:', data),
      error: (err) => console.error('Failed to fetch CSRF token:', err)
    });
  }
}

