import { Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';

@Component({
  selector: 'app-protected-shell',
  imports: [RouterOutlet],
  templateUrl: './protected-shell.html',
  styleUrl: './protected-shell.scss',
})
export class ProtectedShell {}
