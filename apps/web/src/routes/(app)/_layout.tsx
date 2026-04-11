import { Link, Navigate, Outlet, useLocation } from '@tanstack/react-router';
import {
  APP_ROUTES,
  COMPANY_ROUTE_PARAM,
  buildCompanyRoute,
  normalizeCompanySlug,
} from '@petcontrol/shared-constants';
import { isUnauthorizedApiError } from '@/lib/api/rest-client';
import { useCurrentCompanyQuery } from '@/lib/api/domain.queries';
import { useParams } from '@tanstack/react-router';
import { cn } from '@petcontrol/ui/web';
import {
  CalendarRange,
  ClipboardList,
  LogOut,
  Menu,
  MoonStar,
  PawPrint,
  PanelLeftClose,
  PanelLeftOpen,
  SunMedium,
  Users,
} from 'lucide-react';
import { useEffect } from 'react';

import { selectSession, useAuthStore } from '@/lib/auth/auth.store';
import { useUIStore } from '@/stores/ui.store';

export function AppLayout() {
  const hydrated = useAuthStore((state) => state.hydrated);
  const session = useAuthStore(selectSession);
  const clearSession = useAuthStore((state) => state.clearSession);
  const sidebarOpen = useUIStore((state) => state.sidebarOpen);
  const toggleSidebar = useUIStore((state) => state.toggleSidebar);
  const theme = useUIStore((state) => state.theme);
  const setTheme = useUIStore((state) => state.setTheme);
  const params = useParams({ strict: false });
  const location = useLocation();
  const companyQuery = useCurrentCompanyQuery();
  const unauthorizedCompanyContext =
    companyQuery.isError && isUnauthorizedApiError(companyQuery.error);

  useEffect(() => {
    document.documentElement.dataset.theme = theme;
  }, [theme]);

  useEffect(() => {
    if (unauthorizedCompanyContext) {
      clearSession();
    }
  }, [clearSession, unauthorizedCompanyContext]);

  if (hydrated && !session) {
    return <Navigate to={APP_ROUTES.login} replace />;
  }

  if (!hydrated) {
    return <LoadingScreen />;
  }

  if (unauthorizedCompanyContext) {
    return <Navigate to={APP_ROUTES.login} replace />;
  }

  const currentSlug = companyQuery.data?.slug;
  const urlSlug =
    typeof params[COMPANY_ROUTE_PARAM] === 'string'
      ? params[COMPANY_ROUTE_PARAM]
      : undefined;

  if (companyQuery.isError) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center bg-hero-radial px-6 text-center text-white">
        <p className="text-xl font-medium text-rose-400">Erro de Contexto</p>
        <p className="mt-2 text-sm text-slate-400">
          Não conseguimos identificar sua empresa atual.
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

  if (companyQuery.isLoading || !currentSlug || !companyQuery.data) {
    return <LoadingScreen />;
  }

  const company = companyQuery.data;
  const normalizedCurrentSlug = normalizeCompanySlug(currentSlug);
  const normalizedUrlSlug = urlSlug?.toLowerCase();
  const companyDisplayName = company.fantasy_name || company.name;

  if (urlSlug && normalizedUrlSlug !== normalizedCurrentSlug) {
    const segments = location.pathname.split('/');
    if (segments.length > 1) {
      segments[1] = currentSlug;
    }
    const newPathname = segments.length > 1 ? segments.join('/') : buildCompanyRoute(currentSlug, 'dashboard');

    if (location.pathname === newPathname) {
      return <Navigate to={buildCompanyRoute(currentSlug, 'dashboard')} replace />;
    }

    return (
      <Navigate to={newPathname} search={{ ...location.search }} hash={location.hash} replace />
    );
  }

  return (
    <div className="min-h-screen bg-hero-radial text-foreground">
      <div className="mx-auto flex min-h-screen max-w-[1600px] gap-4 p-4 lg:p-6">
        <aside
          className={cn(
            'hidden flex-col rounded-[2rem] border border-white/10 bg-slate-950/80 p-4 shadow-glow backdrop-blur-xl transition-all duration-300 lg:flex',
            sidebarOpen ? 'w-72' : 'w-[5.5rem]',
          )}
        >
          <div className="flex items-center justify-between gap-4 border-b border-white/10 pb-4">
            <div className={cn('space-y-1', !sidebarOpen && 'opacity-0')}>
              <p className="font-display text-2xl text-white">PetControl</p>
              <p className="text-xs uppercase tracking-[0.3em] text-secondary/80">
                tenant aware dashboard
              </p>
            </div>
            <button
              type="button"
              onClick={toggleSidebar}
              title={sidebarOpen ? 'Recolher sidebar' : 'Expandir sidebar'}
              className="rounded-2xl border border-white/10 bg-white/5 p-2 text-white/80 transition hover:bg-white/10"
            >
              {sidebarOpen ? (
                <PanelLeftClose className="h-4 w-4" />
              ) : (
                <PanelLeftOpen className="h-4 w-4" />
              )}
            </button>
          </div>

          <nav className="mt-6 space-y-2 text-sm">
            <SidebarLink
              to={buildCompanyRoute(currentSlug, 'dashboard')}
              icon={Menu}
              label="Dashboard"
              expanded={sidebarOpen}
            />
            <SidebarLink
              to={buildCompanyRoute(currentSlug, 'schedules')}
              icon={CalendarRange}
              label="Schedules"
              expanded={sidebarOpen}
            />
            <SidebarLink
              to={buildCompanyRoute(currentSlug, 'clients')}
              icon={Users}
              label="Clients"
              expanded={sidebarOpen}
            />
            <SidebarLink
              to={buildCompanyRoute(currentSlug, 'pets')}
              icon={PawPrint}
              label="Pets"
              expanded={sidebarOpen}
            />
            <SidebarLink
              to={buildCompanyRoute(currentSlug, 'services')}
              icon={ClipboardList}
              label="Services"
              expanded={sidebarOpen}
            />
          </nav>

          <div className="mt-auto space-y-3 border-t border-white/10 pt-4">
            <button
              type="button"
              onClick={() =>
                setTheme(theme === 'midnight' ? 'ember' : 'midnight')
              }
              className="flex w-full items-center gap-3 rounded-2xl border border-white/10 bg-white/5 px-4 py-3 text-left text-sm text-white/80 transition hover:bg-white/10"
            >
              {theme === 'midnight' ? (
                <MoonStar className="h-4 w-4 text-primary" />
              ) : (
                <SunMedium className="h-4 w-4 text-primary" />
              )}
              {sidebarOpen ? `Tema ${theme}` : null}
            </button>

            <button
              type="button"
              onClick={() => {
                clearSession();
              }}
              className="flex w-full items-center gap-3 rounded-2xl border border-white/10 bg-white/5 px-4 py-3 text-left text-sm text-white/80 transition hover:bg-white/10"
            >
              <LogOut className="h-4 w-4 text-primary" />
              {sidebarOpen ? 'Encerrar sessão' : null}
            </button>
          </div>
        </aside>

        <div className="flex min-w-0 flex-1 flex-col gap-4">
          <header className="rounded-[2rem] border border-white/10 bg-slate-950/75 px-4 py-3 shadow-glow backdrop-blur-xl lg:px-6">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div className="flex items-center gap-3">
                <button
                  type="button"
                  onClick={toggleSidebar}
                  title="Alternar sidebar"
                  className="rounded-2xl border border-white/10 bg-white/5 p-2 text-white/80 transition hover:bg-white/10 lg:hidden"
                >
                  <Menu className="h-4 w-4" />
                </button>
                <div>
                  <p className="text-xs uppercase tracking-[0.3em] text-secondary/80">
                    Operação
                  </p>
                  <h1 className="font-display text-xl text-white">
                    Painel administrativo
                  </h1>
                  <p className="mt-1 text-xs text-slate-400">
                    Tenant atual: {companyDisplayName} ({normalizedCurrentSlug})
                  </p>
                </div>
              </div>

              <div className="flex items-center gap-3 text-sm text-slate-300">
                <span className="rounded-full border border-emerald-400/20 bg-emerald-500/10 px-3 py-1.5 text-emerald-100">
                  @{normalizedCurrentSlug}
                </span>
                <span className="rounded-full border border-white/10 bg-white/5 px-3 py-1.5">
                  {session?.companyId.slice(0, 8)}…
                </span>
                <span className="rounded-full border border-white/10 bg-white/5 px-3 py-1.5">
                  {session?.role}
                </span>
              </div>
            </div>
          </header>

          <main className="min-h-0 flex-1 overflow-auto rounded-[2rem] border border-white/10 bg-white/5 p-4 shadow-glow backdrop-blur-xl lg:p-6">
            <Outlet />
          </main>
        </div>
      </div>
    </div>
  );
}

function SidebarLink({
  to,
  icon: Icon,
  label,
  expanded,
}: {
  to: string;
  icon: typeof Menu;
  label: string;
  expanded: boolean;
}) {
  return (
    <Link
      to={to}
      className="flex items-center gap-3 rounded-2xl px-4 py-3 transition"
      activeProps={{ className: 'bg-primary text-slate-950' }}
      inactiveProps={{
        className: 'bg-white/5 text-white/80 hover:bg-white/10',
      }}
    >
      <Icon className="h-4 w-4" />
      {expanded ? label : null}
    </Link>
  );
}

function LoadingScreen() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-hero-radial px-6 text-center text-white">
      <div className="max-w-sm rounded-[2rem] border border-white/10 bg-slate-950/80 px-8 py-10 shadow-glow backdrop-blur-xl">
        <p className="font-display text-3xl">Carregando painel</p>
        <p className="mt-3 text-sm text-slate-300">
          Sincronizando a sessão persistida e preparando a navegação.
        </p>
      </div>
    </div>
  );
}
