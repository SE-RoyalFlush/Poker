import { Routes } from '@angular/router';
import { authGuard } from './core/auth.guard';

export const routes: Routes = [
  { path: '', pathMatch: 'full', redirectTo: 'home' },

  {
    path: '',
    data: { access: 'public' },
    children: [
      {
        path: 'home',
        loadComponent: () => import('./pages/home/home.component').then((m) => m.HomeComponent),
      },
      {
        path: 'login',
        loadComponent: () =>
          import('./features/auth/login/login.component').then((m) => m.LoginComponent),
      },
      {
        path: 'register',
        loadComponent: () =>
          import('./features/auth/register/register.component').then((m) => m.RegisterComponent),
      },
      {
        path: 'admin',
        loadComponent: () => import('./pages/admin/admin').then((m) => m.AdminComponent),
        canActivate: [authGuard],
      },
      {
        path: 'dashboard',
        loadComponent: () => import('./pages/dashboard/dashboard').then((m) => m.DashboardComponent),
        canActivate: [authGuard],
      },
      {
        path: 'profile/:id',
        loadComponent: () => import('./pages/profile/profile').then((m) => m.ProfileComponent),
        canActivate: [authGuard],
      },
      {
        path: 'leaderboard',
        loadComponent: () => import('./pages/leaderboard/leaderboard').then((m) => m.LeaderboardComponent),
      },
    ],
  },

  {
    path: '',
    data: { access: 'protected' },
    loadComponent: () =>
      import('./layouts/protected-shell/protected-shell').then((m) => m.ProtectedShell),
    children: [
      {
        path: 'lobby/:code',
        loadComponent: () => import('./pages/lobby/lobby').then((m) => m.Lobby),
        canActivate: [authGuard],
      },
      {
        path: 'table/:id',
        loadComponent: () => import('./pages/table/table').then((m) => m.Table),
        canActivate: [authGuard],
      },
    ],
  },

  { path: '**', redirectTo: 'home' },
];
