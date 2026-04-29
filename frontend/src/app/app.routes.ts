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
        loadComponent: () => import('./pages/login/login').then((m) => m.Login),
      },
      {
        path: 'register',
        loadComponent: () =>
          import('./features/auth/register/register.component').then((m) => m.RegisterComponent),
      },
      {
        path: 'admin',
        loadComponent: () => import('./pages/admin/admin').then((m) => m.AdminComponent),
      },
      {
        path: 'dashboard',
        loadComponent: () => import('./pages/dashboard/dashboard').then((m) => m.DashboardComponent),
        canActivate: [authGuard],
      },
    ],
  },

  {
    path: '',
    data: { access: 'protected' }, // guard will be added in auth story
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
