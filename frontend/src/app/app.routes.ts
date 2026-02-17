import { Routes } from '@angular/router';

export const routes: Routes = [
  { path: '', pathMatch: 'full', redirectTo: 'login' },

  {
    path: '',
    data: { access: 'public' },
    children: [
      {
        path: 'login',
        loadComponent: () => import('./pages/login/login').then((m) => m.Login),
      },
      {
        path: 'register',
        loadComponent: () => import('./pages/register/register').then((m) => m.Register),
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
        path: 'dashboard',
        loadComponent: () => import('./pages/dashboard/dashboard').then((m) => m.Dashboard),
      },
      {
        path: 'lobby',
        loadComponent: () => import('./pages/lobby/lobby').then((m) => m.Lobby),
      },
      {
        path: 'table/:id',
        loadComponent: () => import('./pages/table/table').then((m) => m.Table),
      },
    ],
  },

  { path: '**', redirectTo: 'login' },
];
