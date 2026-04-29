import { routes } from './app.routes';

describe('app routes', () => {
  it('should define the canonical lobby route used by room navigation flows', () => {
    const allChildren = routes.flatMap((route) => route.children ?? []);
    const lobbyRoute = allChildren.find((route) => route.path === 'lobby/:code');

    expect(lobbyRoute).toBeTruthy();
  });

  it('should not keep the legacy room route alongside the lobby contract', () => {
    const allPaths = routes.flatMap((route) => route.children?.map((child) => child.path) ?? []);

    expect(allPaths).not.toContain('room/:code');
  });
});
