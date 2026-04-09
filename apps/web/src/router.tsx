/* eslint-disable react-refresh/only-export-components */
import { QueryClient } from '@tanstack/react-query';
import { APP_ROUTES, APP_ROUTE_SEGMENTS } from '@petcontrol/shared-constants';
import {
  Navigate,
  Outlet,
  createRoute,
  createRootRoute,
  createRouter,
} from '@tanstack/react-router';

import { AppLayout } from '@/routes/(app)/_layout';
import { DashboardPage } from '@/routes/(app)/dashboard';
import { SchedulesPage } from '@/routes/(app)/schedules';
import { LoginPage } from '@/routes/(auth)/login';
import { useAuthStore } from '@/lib/auth/auth.store';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
    mutations: {
      retry: 0,
    },
  },
});

const rootRoute = createRootRoute({
  component: RootRoute,
});

const homeRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: APP_ROUTES.home,
  component: HomeRedirect,
});

const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: APP_ROUTE_SEGMENTS.login,
  component: LoginPage,
});

const appRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: 'app',
  component: AppLayout,
});

const dashboardRoute = createRoute({
  getParentRoute: () => appRoute,
  path: APP_ROUTE_SEGMENTS.dashboard,
  component: DashboardPage,
});

const schedulesRoute = createRoute({
  getParentRoute: () => appRoute,
  path: APP_ROUTE_SEGMENTS.schedules,
  component: SchedulesPage,
});

const routeTree = rootRoute.addChildren([
  homeRoute,
  loginRoute,
  appRoute.addChildren([dashboardRoute, schedulesRoute]),
]);

export const router = createRouter({
  routeTree,
  defaultPreload: 'intent',
  context: {
    queryClient,
  },
});

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}

export function queryClientForWeb() {
  return queryClient;
}

function RootRoute() {
  return <Outlet />;
}

function HomeRedirect() {
  const hydrated = useAuthStore((state) => state.hydrated);
  const session = useAuthStore((state) => state.session);

  if (!hydrated) {
    return <SplashScreen />;
  }

  return (
    <Navigate to={session ? APP_ROUTES.dashboard : APP_ROUTES.login} replace />
  );
}

function SplashScreen() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-hero-radial px-6 text-center text-white">
      <div className="max-w-sm rounded-[2rem] border border-white/10 bg-slate-950/80 px-8 py-10 shadow-glow backdrop-blur-xl">
        <p className="font-display text-3xl">PetControl</p>
        <p className="mt-3 text-sm text-slate-300">
          Inicializando o frontend e carregando a sessão persistida.
        </p>
      </div>
    </div>
  );
}
