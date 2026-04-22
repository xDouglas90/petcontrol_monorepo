import {
  Link,
  Navigate,
  Outlet,
  useLocation,
  useNavigate,
} from '@tanstack/react-router';
import {
  APP_ROUTES,
  COMPANY_ROUTE_PARAM,
  buildCompanyRoute,
  normalizeCompanySlug,
} from '@petcontrol/shared-constants';
import { isUnauthorizedApiError } from '@/lib/api/rest-client';
import {
  useCurrentCompanyQuery,
  useCurrentUserQuery,
} from '@/lib/api/domain.queries';
import { useParams } from '@tanstack/react-router';
import { cn } from '@petcontrol/ui/web';
import {
  CalendarRange,
  ChevronRight,
  Cog,
  LayoutGrid,
  LogOut,
  PawPrint,
  Sparkles,
  Users,
  X,
} from 'lucide-react';
import type { ReactNode } from 'react';
import { useEffect, useState } from 'react';

import { selectSession, useAuthStore } from '@/lib/auth/auth.store';
import { useUIStore } from '@/stores/ui.store';
import { AdminSupportChatAside } from '@/components/admin-support-chat-aside';

const PLAN_UPGRADE_FLOW = {
  trial: 'starter',
  starter: 'basic',
  basic: 'essential',
  essential: 'premium',
} as const;

export function AppLayout() {
  const navigate = useNavigate();
  const [isDesktopViewport, setIsDesktopViewport] = useState(() => {
    if (
      typeof window === 'undefined' ||
      typeof window.matchMedia !== 'function'
    ) {
      return true;
    }

    return window.matchMedia('(min-width: 1024px)').matches;
  });
  const hydrated = useAuthStore((state) => state.hydrated);
  const session = useAuthStore(selectSession);
  const clearSession = useAuthStore((state) => state.clearSession);
  const sidebarOpen = useUIStore((state) => state.sidebarOpen);
  const toggleSidebar = useUIStore((state) => state.toggleSidebar);
  const setSidebarOpen = useUIStore((state) => state.setSidebarOpen);
  const params = useParams({ strict: false });
  const location = useLocation();
  const companyQuery = useCurrentCompanyQuery();
  const currentUserQuery = useCurrentUserQuery();
  const unauthorizedCompanyContext =
    companyQuery.isError && isUnauthorizedApiError(companyQuery.error);
  const unauthorizedUserContext =
    currentUserQuery.isError && isUnauthorizedApiError(currentUserQuery.error);

  useEffect(() => {
    if (
      typeof window === 'undefined' ||
      typeof window.matchMedia !== 'function'
    ) {
      return;
    }

    const desktopQuery = window.matchMedia('(min-width: 1024px)');

    const syncSidebarForViewport = (matches: boolean) => {
      setIsDesktopViewport(matches);
      setSidebarOpen(matches);
    };

    syncSidebarForViewport(desktopQuery.matches);

    const handleViewportChange = (event: MediaQueryListEvent) => {
      syncSidebarForViewport(event.matches);
    };

    desktopQuery.addEventListener('change', handleViewportChange);

    return () => {
      desktopQuery.removeEventListener('change', handleViewportChange);
    };
  }, [setSidebarOpen]);

  useEffect(() => {
    if (unauthorizedCompanyContext || unauthorizedUserContext) {
      clearSession();
    }
  }, [clearSession, unauthorizedCompanyContext, unauthorizedUserContext]);

  const shouldRedirectToLogin =
    hydrated &&
    (!session || unauthorizedCompanyContext || unauthorizedUserContext);

  useEffect(() => {
    if (!shouldRedirectToLogin || location.pathname === APP_ROUTES.login) {
      return;
    }

    void navigate({
      to: APP_ROUTES.login,
      replace: true,
    });
  }, [location.pathname, navigate, shouldRedirectToLogin]);

  if (!hydrated) {
    return <LoadingScreen />;
  }

  if (shouldRedirectToLogin) {
    return <LoadingScreen />;
  }

  const currentSlug = companyQuery.data?.slug;
  const urlSlug =
    typeof params[COMPANY_ROUTE_PARAM] === 'string'
      ? params[COMPANY_ROUTE_PARAM]
      : undefined;

  if (companyQuery.isError || currentUserQuery.isError) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center bg-[#ebe8e4] px-6 text-center text-stone-900">
        <p className="text-xl font-medium text-rose-600">Erro de Contexto</p>
        <p className="mt-2 text-sm text-stone-500">
          Não conseguimos carregar sua empresa ou o perfil autenticado.
        </p>
        <div className="mt-6 flex gap-4">
          <button
            onClick={() => {
              void companyQuery.refetch();
              void currentUserQuery.refetch();
            }}
            className="rounded-xl bg-sky-600 px-4 py-2 text-sm text-white transition hover:bg-sky-700"
          >
            Tentar novamente
          </button>
          <button
            onClick={() => clearSession()}
            className="rounded-xl border border-rose-200 bg-rose-50 px-4 py-2 text-sm text-rose-600 transition hover:bg-rose-100"
          >
            Sair
          </button>
        </div>
      </div>
    );
  }

  if (
    companyQuery.isLoading ||
    currentUserQuery.isLoading ||
    !currentSlug ||
    !companyQuery.data
  ) {
    return <LoadingScreen />;
  }

  const company = companyQuery.data;
  const currentUser = currentUserQuery.data;
  const normalizedCurrentSlug = normalizeCompanySlug(currentSlug);
  const normalizedUrlSlug = urlSlug?.toLowerCase();
  const companyDisplayName = company.fantasy_name || company.name;
  const canViewSettings =
    currentUser?.settings_access?.can_view ?? currentUser?.role === 'admin';

  if (
    urlSlug &&
    normalizedCurrentSlug &&
    normalizedUrlSlug !== normalizedCurrentSlug
  ) {
    const segments = location.pathname.split('/');
    if (segments.length > 1) {
      segments[1] = normalizedCurrentSlug;
    }
    const targetPath =
      segments.length > 1
        ? segments.join('/')
        : buildCompanyRoute(normalizedCurrentSlug, 'dashboard');

    if (location.pathname.toLowerCase() === targetPath.toLowerCase()) {
      return null;
    }

    return (
      <Navigate
        to={targetPath}
        search={{ ...location.search }}
        hash={location.hash}
        replace
      />
    );
  }

  const suggestedPlan = resolveSuggestedPlan(company.active_package);

  const handleSidebarLinkClick = () => {
    if (!isDesktopViewport) {
      setSidebarOpen(false);
    }
  };

  return (
    <div className="min-h-screen bg-white text-stone-900">
      {!isDesktopViewport && sidebarOpen ? (
        <button
          type="button"
          aria-label="Fechar menu lateral"
          onClick={() => setSidebarOpen(false)}
          className="fixed inset-0 z-40 bg-stone-950/25 backdrop-blur-[2px] lg:hidden"
        />
      ) : null}

      <div className="mx-auto flex min-h-screen max-w-[1920px]">
        <aside
          className={cn(
            'flex-col bg-white border-r border-stone-100 transition-all duration-300',
            isDesktopViewport
              ? cn('hidden lg:flex', sidebarOpen ? 'w-[19.5rem]' : 'w-[5rem]')
              : cn(
                  'fixed inset-y-3 left-3 z-50 flex w-[min(21rem,calc(100vw-1.5rem))] transform',
                  sidebarOpen
                    ? 'translate-x-0 opacity-100'
                    : '-translate-x-[110%] opacity-0 pointer-events-none',
                ),
          )}
          aria-hidden={!isDesktopViewport && !sidebarOpen ? 'true' : undefined}
        >
          <div className="flex items-center justify-between border-b border-stone-100 px-5 py-5">
            <div
              className={cn(
                'flex items-center gap-3 overflow-hidden transition-all duration-300',
                isDesktopViewport && !sidebarOpen
                  ? 'w-0 opacity-0'
                  : 'w-auto opacity-100',
              )}
            >
              <TenantBrand
                companyName={companyDisplayName}
                logoUrl={company.logo_url}
              />
              <div className="min-w-0">
                <p className="truncate font-display text-xl text-stone-900">
                  {companyDisplayName}
                </p>
                <p className="truncate text-xs uppercase tracking-[0.28em] text-stone-400">
                  admin workspace
                </p>
              </div>
            </div>

            <button
              type="button"
              onClick={toggleSidebar}
              title={sidebarOpen ? 'Recolher sidebar' : 'Expandir sidebar'}
              className="inline-flex h-11 w-11 items-center justify-center rounded-2xl border border-stone-200 bg-stone-50 text-stone-500 transition hover:border-stone-300 hover:bg-stone-100 hover:text-stone-900"
            >
              {isDesktopViewport ? (
                <ChevronRight
                  className={cn(
                    'h-4 w-4 transition-transform duration-300',
                    sidebarOpen ? 'rotate-180' : 'rotate-0',
                  )}
                />
              ) : (
                <X className="h-4 w-4" />
              )}
            </button>
          </div>

          <div className="h-4" />

          <nav className="mt-7 space-y-1 px-4 pb-4">
            <SidebarLink
              to={buildCompanyRoute(currentSlug, 'dashboard')}
              icon={LayoutGrid}
              label="Dashboard"
              expanded={isDesktopViewport ? sidebarOpen : true}
              collapsedDesktop={isDesktopViewport && !sidebarOpen}
              onNavigate={handleSidebarLinkClick}
            />
            <SidebarLink
              to={buildCompanyRoute(currentSlug, 'schedules')}
              icon={CalendarRange}
              label="Agendamentos"
              expanded={isDesktopViewport ? sidebarOpen : true}
              collapsedDesktop={isDesktopViewport && !sidebarOpen}
              onNavigate={handleSidebarLinkClick}
            />
            <SidebarLink
              to={buildCompanyRoute(currentSlug, 'clients')}
              icon={Users}
              label="Clientes"
              expanded={isDesktopViewport ? sidebarOpen : true}
              collapsedDesktop={isDesktopViewport && !sidebarOpen}
              onNavigate={handleSidebarLinkClick}
            />
            <SidebarLink
              to={buildCompanyRoute(currentSlug, 'pets')}
              icon={PawPrint}
              label="Pets"
              expanded={isDesktopViewport ? sidebarOpen : true}
              collapsedDesktop={isDesktopViewport && !sidebarOpen}
              onNavigate={handleSidebarLinkClick}
            />
          </nav>

          <div className="px-4">
            <UpgradeCard
              activePackage={company.active_package}
              expanded={isDesktopViewport ? sidebarOpen : true}
              suggestedPlan={suggestedPlan}
            />
          </div>

          <div className="mt-auto space-y-2 border-t border-stone-100 px-4 py-4">
            {canViewSettings ? (
              <SidebarLink
                to={`/${normalizeCompanySlug(currentSlug)}/settings`}
                icon={Cog}
                label="Configurações"
                expanded={isDesktopViewport ? sidebarOpen : true}
                collapsedDesktop={isDesktopViewport && !sidebarOpen}
                onNavigate={handleSidebarLinkClick}
              />
            ) : null}

            <button
              type="button"
              onClick={() => {
                clearSession();
              }}
              className={cn(
                'flex w-full items-center rounded-2xl text-sm text-stone-500 transition hover:bg-stone-100 hover:text-stone-900',
                isDesktopViewport && !sidebarOpen
                  ? 'justify-center px-0 py-3'
                  : 'gap-3 px-4 py-3',
              )}
            >
              <LogOut className="h-4 w-4" />
              <SidebarLabel expanded={isDesktopViewport ? sidebarOpen : true}>
                Sair
              </SidebarLabel>
            </button>
          </div>
        </aside>

        <div className="flex min-w-0 flex-1 flex-col gap-4">
          <header className="flex items-center justify-between rounded-[2rem] border border-white/70 bg-white/80 px-4 py-4 shadow-[0_18px_45px_rgba(15,23,42,0.06)] backdrop-blur-xl lg:hidden">
            <button
              type="button"
              onClick={toggleSidebar}
              title="Alternar sidebar"
              className="inline-flex h-11 w-11 items-center justify-center rounded-2xl border border-stone-200 bg-stone-50 text-stone-500 transition hover:border-stone-300 hover:bg-stone-100 hover:text-stone-900"
            >
              <LayoutGrid className="h-4 w-4" />
            </button>
            <div className="flex items-center gap-3">
              <UserAvatar
                shortName={currentUser?.short_name ?? null}
                imageUrl={currentUser?.image_url ?? null}
                size="sm"
              />
            </div>
          </header>

          <main className="min-h-0 flex-1">
            <Outlet />
          </main>
        </div>

        {currentUser?.role === 'admin' ? (
          <AdminSupportChatAside className="hidden xl:flex" />
        ) : null}
      </div>
    </div>
  );
}

function SidebarLink({
  to,
  icon: Icon,
  label,
  expanded,
  collapsedDesktop,
  onNavigate,
}: {
  to: string;
  icon: typeof LayoutGrid;
  label: string;
  expanded: boolean;
  collapsedDesktop?: boolean;
  onNavigate?: () => void;
}) {
  return (
    <Link
      to={to}
      onClick={onNavigate}
      className={cn(
        'group flex items-center rounded-2xl py-3 text-sm transition',
        collapsedDesktop ? 'justify-center px-0' : 'gap-3 px-4',
      )}
      activeProps={{
        className: cn(
          'bg-sky-100 text-sky-600 font-bold shadow-sm',
          collapsedDesktop ? 'justify-center px-0 py-3' : 'gap-3 px-4 py-3',
        ),
      }}
      inactiveProps={{
        className: cn(
          'text-stone-500 hover:bg-sky-50 hover:text-sky-600',
          collapsedDesktop ? 'justify-center px-0 py-3' : 'gap-3 px-4 py-3',
        ),
      }}
    >
      <Icon className="h-4 w-4 shrink-0" />
      <SidebarLabel expanded={expanded}>{label}</SidebarLabel>
    </Link>
  );
}

function SidebarLabel({
  expanded,
  children,
}: {
  expanded: boolean;
  children: ReactNode;
}) {
  return (
    <span
      className={cn(
        'overflow-hidden whitespace-nowrap transition-[max-width,opacity,transform] duration-200 ease-out',
        expanded
          ? 'max-w-[14rem] translate-x-0 opacity-100 delay-100'
          : 'max-w-0 -translate-x-1 opacity-0 delay-0',
      )}
      aria-hidden={!expanded ? 'true' : undefined}
    >
      {children}
    </span>
  );
}

function TenantBrand({
  companyName,
  logoUrl,
}: {
  companyName: string;
  logoUrl?: string | null;
}) {
  if (logoUrl) {
    return (
      <div className="flex h-11 w-11 items-center justify-center overflow-hidden rounded-2xl border border-stone-200 bg-stone-50">
        <img
          src={logoUrl}
          alt={`Logo de ${companyName}`}
          className="h-full w-full object-cover"
        />
      </div>
    );
  }

  return (
    <div className="flex h-11 w-11 items-center justify-center rounded-2xl bg-sky-600 font-display text-sm uppercase tracking-[0.18em] text-white">
      {resolveInitials(companyName)}
    </div>
  );
}

function UserAvatar({
  shortName,
  imageUrl,
  size = 'md',
}: {
  shortName?: string | null;
  imageUrl?: string | null;
  size?: 'sm' | 'md';
}) {
  const sizeClass = size === 'sm' ? 'h-10 w-10 text-xs' : 'h-12 w-12 text-sm';

  if (imageUrl) {
    return (
      <div
        className={cn(
          'overflow-hidden rounded-2xl border border-white/20 bg-white/15',
          sizeClass,
        )}
      >
        <img
          src={imageUrl}
          alt={shortName ? `Avatar de ${shortName}` : 'Avatar do usuário'}
          className="h-full w-full object-cover"
        />
      </div>
    );
  }

  return (
    <div
      className={cn(
        'flex items-center justify-center rounded-2xl bg-sky-600 font-display uppercase tracking-[0.18em] text-white',
        sizeClass,
      )}
    >
      {resolveInitials(shortName || 'Usuário')}
    </div>
  );
}

function UpgradeCard({
  activePackage,
  expanded,
  suggestedPlan,
}: {
  activePackage: string;
  expanded: boolean;
  suggestedPlan: string | null;
}) {
  const isSpecialTier =
    activePackage === 'premium' || activePackage === 'internal';

  if (!expanded) {
    return (
      <div className="flex justify-center py-4">
        <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-sky-50 text-sky-500 shadow-sm border border-sky-100">
          <Sparkles className="h-5 w-5" />
        </div>
      </div>
    );
  }

  const title = isSpecialTier
    ? 'Plano consolidado'
    : `Upgrade para ${suggestedPlan}`;

  return (
    <div className="mt-4 rounded-[2rem] bg-stone-50 p-6 border border-stone-100">
      <div className="flex flex-col items-center text-center">
        <h4 className="font-display text-lg text-stone-900">{title}</h4>
        <p className="mt-2 text-xs leading-relaxed text-stone-400 px-2">
          {isSpecialTier
            ? 'Você já possui todos os recursos liberados.'
            : 'Ganhe 1 mês grátis e desbloqueie novos recursos agora.'}
        </p>

        <button
          type="button"
          className={cn(
            'mt-6 inline-flex w-full items-center justify-center rounded-2xl bg-sky-100 px-4 py-3 text-sm font-bold text-sky-600 transition hover:bg-sky-200 shadow-sm',
            isSpecialTier && 'bg-sky-600 text-white hover:bg-sky-700',
          )}
        >
          {isSpecialTier ? 'Ver detalhes' : 'Fazer Upgrade'}
        </button>
      </div>
    </div>
  );
}

function LoadingScreen() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-[#ebe8e4] px-6 text-center text-stone-900">
      <div className="max-w-sm rounded-[2rem] border border-white/70 bg-white/90 px-8 py-10 shadow-[0_18px_45px_rgba(15,23,42,0.08)]">
        <p className="font-display text-3xl">PetControl</p>
        <p className="mt-3 text-sm text-stone-500">
          Sincronizando o novo shell e carregando o contexto autenticado.
        </p>
      </div>
    </div>
  );
}

function resolveInitials(value: string) {
  return value
    .trim()
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part[0])
    .join('')
    .toUpperCase();
}

function resolveSuggestedPlan(activePackage: string) {
  return (
    PLAN_UPGRADE_FLOW[activePackage as keyof typeof PLAN_UPGRADE_FLOW] ?? null
  );
}
