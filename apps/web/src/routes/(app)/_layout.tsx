import { Link, Navigate, Outlet } from '@tanstack/react-router';
import { APP_ROUTES } from '@petcontrol/shared-constants';
import { cn } from '@petcontrol/ui/web';
import {
  LogOut,
  Menu,
  MoonStar,
  PanelLeftClose,
  PanelLeftOpen,
  SunMedium,
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

  useEffect(() => {
    document.documentElement.dataset.theme = theme;
  }, [theme]);

  if (hydrated && !session) {
    return <Navigate to={APP_ROUTES.login} replace />;
  }

  if (!hydrated) {
    return <LoadingScreen />;
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
              to={APP_ROUTES.dashboard}
              icon={Menu}
              label="Dashboard"
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
                </div>
              </div>

              <div className="flex items-center gap-3 text-sm text-slate-300">
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
  to: typeof APP_ROUTES.dashboard;
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
