import { routes } from './app.routes';

describe('app routes', () => {
  it('should define the canonical lobby route used by room navigation flows', () => {
    const allChildren = routes.flatMap((route) => route.children ?? []);
    const lobbyRoute = allChildren.find((route) => route.path === 'lobby/:code');

    expect(lobbyRoute).toBeTruthy();
    expect(lobbyRoute?.canActivate?.length).toBeGreaterThan(0);
  });

  it('should not keep the legacy room route alongside the lobby contract', () => {
    const allPaths = routes.flatMap((route) => route.children?.map((child) => child.path) ?? []);

    expect(allPaths).not.toContain('room/:code');
  });

  it('should keep the protected-shell room flow routes behind authGuard', () => {
    const allChildren = routes.flatMap((route) => route.children ?? []);
    const protectedPaths = new Map(
      allChildren
        .filter((route) => route.path === 'lobby/:code' || route.path === 'table/:id')
        .map((route) => [route.path, route.canActivate?.length ?? 0])
    );

    expect(protectedPaths.get('lobby/:code')).toBeGreaterThan(0);
    expect(protectedPaths.get('table/:id')).toBeGreaterThan(0);
  });
});
