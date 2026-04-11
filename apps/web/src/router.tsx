/* eslint-disable react-refresh/only-export-components */
import { QueryClient } from '@tanstack/react-query';
import {
  Navigate,
  Outlet,
  createRoute,
  createRootRoute,
  createRouter,
} from '@tanstack/react-router';
import {
  APP_ROUTES,
  APP_ROUTE_SEGMENTS,
  buildCompanyRoute,
} from '@petcontrol/shared-constants';

import { AppLayout } from '@/routes/(app)/_layout';
import { ClientsPage } from '@/routes/(app)/clients';
import { DashboardPage } from '@/routes/(app)/dashboard';
import { PetsPage } from '@/routes/(app)/pets';
import { SchedulesPage } from '@/routes/(app)/schedules';
import { ServicesPage } from '@/routes/(app)/services';
import { LoginPage } from '@/routes/(auth)/login';
import { isUnauthorizedApiError } from '@/lib/api/rest-client';
import { useCurrentCompanyQuery } from '@/lib/api/domain.queries';
import { useAuthStore } from '@/lib/auth/auth.store';
import { useEffect } from 'react';

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
  path: '/',
  component: HomeRedirect,
});

const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: APP_ROUTE_SEGMENTS.login,
  component: LoginPage,
});

const companyRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '$companySlug',
  component: AppLayout,
});

const dashboardRoute = createRoute({
  getParentRoute: () => companyRoute,
  path: APP_ROUTE_SEGMENTS.dashboard,
  component: DashboardPage,
});

const schedulesRoute = createRoute({
  getParentRoute: () => companyRoute,
  path: APP_ROUTE_SEGMENTS.schedules,
  component: SchedulesPage,
});

const clientsRoute = createRoute({
  getParentRoute: () => companyRoute,
  path: APP_ROUTE_SEGMENTS.clients,
  component: ClientsPage,
});

const petsRoute = createRoute({
  getParentRoute: () => companyRoute,
  path: APP_ROUTE_SEGMENTS.pets,
  component: PetsPage,
});

const servicesRoute = createRoute({
  getParentRoute: () => companyRoute,
  path: APP_ROUTE_SEGMENTS.services,
  component: ServicesPage,
});

const routeTree = rootRoute.addChildren([
  homeRoute,
  loginRoute,
  companyRoute.addChildren([
    dashboardRoute,
    schedulesRoute,
    clientsRoute,
    petsRoute,
    servicesRoute,
  ]),
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
  const clearSession = useAuthStore((state) => state.clearSession);
  const companyQuery = useCurrentCompanyQuery();
  const unauthorizedCompanyContext =
    companyQuery.isError && isUnauthorizedApiError(companyQuery.error);

  useEffect(() => {
    if (unauthorizedCompanyContext) {
      clearSession();
    }
  }, [clearSession, unauthorizedCompanyContext]);

  if (!hydrated) {
    return <SplashScreen />;
  }

  if (!session) {
    return <Navigate to={APP_ROUTES.login} replace />;
  }

  if (unauthorizedCompanyContext) {
    return <Navigate to={APP_ROUTES.login} replace />;
  }

  if (companyQuery.isError) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center bg-hero-radial px-6 text-center text-white">
        <p className="text-xl font-medium text-rose-400">
          Falha ao carregar contexto
        </p>
        <p className="mt-2 text-sm text-slate-400">
          Não foi possível recuperar os dados da sua empresa.
        </p>
        <div className="mt-6 flex gap-4">
          <button
            onClick={() => void companyQuery.refetch()}
            className="rounded-xl bg-white/10 px-4 py-2 text-sm hover:bg-white/20"
          >
            Tentar novamente
          </button>
          <button
            onClick={() => clearSession()}
            className="rounded-xl border border-rose-500/30 bg-rose-500/10 px-4 py-2 text-sm text-rose-400 hover:bg-rose-500/20"
          >
            Sair
          </button>
        </div>
      </div>
    );
  }

  if (companyQuery.data?.slug) {
    return (
      <Navigate
        to={buildCompanyRoute(companyQuery.data.slug, 'dashboard')}
        replace
      />
    );
  }

  return <SplashScreen />;
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
