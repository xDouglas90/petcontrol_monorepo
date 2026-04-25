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
  ContactRound,
  Sparkles,
  Users,
  X,
  Plus,
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
      <div className="flex app-shell flex-col items-center justify-center px-6 text-center">
        <p className="text-xl font-medium text-rose-500">Erro de Contexto</p>
        <p className="mt-2 text-sm text-muted">
          Não conseguimos carregar sua empresa ou o perfil autenticado.
        </p>
        <div className="mt-6 flex gap-4">
          <button
            onClick={() => {
              void companyQuery.refetch();
              void currentUserQuery.refetch();
            }}
            className="rounded-xl bg-primary px-4 py-2 text-sm text-stone-900 font-bold transition hover:bg-primary/90"
          >
            Tentar novamente
          </button>
          <button
            onClick={() => clearSession()}
            className="rounded-xl border border-border/50 bg-surface/50 px-4 py-2 text-sm text-rose-500 transition hover:bg-surface"
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
  const canViewPeople =
    currentUser?.role === 'admin' || currentUser?.role === 'system';
  const isDashboardRoute = location.pathname.endsWith('/dashboard');

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
    <div className="app-shell">
      {!isDesktopViewport && sidebarOpen ? (
        <button
          type="button"
          aria-label="Fechar menu lateral"
          onClick={() => setSidebarOpen(false)}
          className="fixed inset-0 z-40 bg-stone-950/50 backdrop-blur-[2px] lg:hidden"
        />
      ) : null}

      <div className="mx-auto flex min-h-screen max-w-[1920px]">
        <aside
          className={cn(
            'flex flex-col overflow-hidden transition-all duration-300 lg:sticky lg:top-0 lg:h-screen lg:max-h-screen app-sidebar',
            isDesktopViewport
              ? cn('hidden lg:flex', sidebarOpen ? 'w-[19.5rem]' : 'w-[5rem]')
              : cn(
                  'fixed inset-y-3 left-3 z-50 flex w-[min(21rem,calc(100vw-1.5rem))] transform',
                  sidebarOpen
                    ? 'translate-x-0 opacity-100'
                    : '-translate-x-[110%] opacity-0 pointer-events-none',
                ),
          )}
          aria-hidden={!isDesktopViewport && !sidebarOpen ? true : undefined}
        >
          <div className="flex items-center justify-between border-b border-border/50 px-5 py-5">
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
                <p className="truncate font-display text-xl text-foreground">
                  {companyDisplayName}
                </p>
                <p className="app-eyebrow truncate">
                  admin workspace
                </p>
              </div>
            </div>

            <button
              type="button"
              onClick={toggleSidebar}
              title={sidebarOpen ? 'Recolher sidebar' : 'Expandir sidebar'}
              className="inline-flex h-11 w-11 items-center justify-center rounded-2xl border border-border/50 bg-surface/50 text-muted transition hover:border-border hover:bg-surface hover:text-foreground"
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

          <div className="flex min-h-0 flex-1 flex-col">
            <div className="flex-1 overflow-y-auto">
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
                {canViewPeople ? (
                  <SidebarLink
                    to={buildCompanyRoute(currentSlug, 'people')}
                    icon={ContactRound}
                    label={
                      <span className="flex w-full items-center">
                        <span className="flex-grow truncate">Pessoas</span>
                        <button
                          aria-label="Adicionar pessoa"
                          type="button"
                          title="Adicionar pessoa"
                          className="ml-2 inline-flex h-6 w-6 items-center justify-center rounded-full border border-primary/30 bg-primary/10 text-xs font-bold text-primary shadow hover:bg-primary/20 focus:outline-none focus:ring-2 focus:ring-primary/50"
                          onClick={(e) => {
                            e.stopPropagation();
                            e.preventDefault();
                            navigate({
                              to: buildCompanyRoute(currentSlug, 'people'),
                              search: {},
                              hash: '',
                              replace: false,
                            });
                            setTimeout(() => {
                              const event = new CustomEvent(
                                'open-people-create-form',
                              );
                              window.dispatchEvent(event);
                            }, 50);
                          }}
                        >
                          <Plus className="h-4 w-4" />
                        </button>
                      </span>
                    }
                    expanded={isDesktopViewport ? sidebarOpen : true}
                    collapsedDesktop={isDesktopViewport && !sidebarOpen}
                    onNavigate={handleSidebarLinkClick}
                  />
                ) : null}
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
                  label={
                    <span className="flex w-full items-center">
                      <span className="flex-grow truncate">Pets</span>
                      <button
                        aria-label="Adicionar pet"
                        type="button"
                        title="Adicionar pet"
                        className="ml-2 inline-flex h-6 w-6 items-center justify-center rounded-full border border-primary/30 bg-primary/10 text-xs font-bold text-primary shadow hover:bg-primary/20 focus:outline-none focus:ring-2 focus:ring-primary/50"
                        onClick={(e) => {
                          e.stopPropagation();
                          e.preventDefault();
                          const event = new CustomEvent(
                            'open-pets-create-form',
                          );
                          window.dispatchEvent(event);
                        }}
                      >
                        <Plus className="h-4 w-4" />
                      </button>
                    </span>
                  }
                  expanded={isDesktopViewport ? sidebarOpen : true}
                  collapsedDesktop={isDesktopViewport && !sidebarOpen}
                  onNavigate={handleSidebarLinkClick}
                />
                <SidebarLink
                  to={buildCompanyRoute(currentSlug, 'services')}
                  icon={Sparkles}
                  label={
                    <span className="flex w-full items-center">
                      <span className="flex-grow truncate">Serviços</span>
                      <button
                        aria-label="Adicionar serviço"
                        type="button"
                        title="Adicionar serviço"
                        className="ml-2 inline-flex h-6 w-6 items-center justify-center rounded-full border border-primary/30 bg-primary/10 text-xs font-bold text-primary shadow hover:bg-primary/20 focus:outline-none focus:ring-2 focus:ring-primary/50"
                        onClick={(e) => {
                          e.stopPropagation();
                          e.preventDefault();
                          navigate({
                            to: buildCompanyRoute(currentSlug, 'services'),
                            search: {},
                            hash: '',
                            replace: false,
                          });
                          setTimeout(() => {
                            const event = new CustomEvent(
                              'open-services-create-form',
                            );
                            window.dispatchEvent(event);
                          }, 50);
                        }}
                      >
                        <Plus className="h-4 w-4" />
                      </button>
                    </span>
                  }
                  expanded={isDesktopViewport ? sidebarOpen : true}
                  collapsedDesktop={isDesktopViewport && !sidebarOpen}
                  onNavigate={handleSidebarLinkClick}
                />
              </nav>

              <div className="px-4 pb-4">
                <UpgradeCard
                  activePackage={company.active_package}
                  expanded={isDesktopViewport ? sidebarOpen : true}
                  suggestedPlan={suggestedPlan}
                />
              </div>
            </div>

            <div className="mt-auto w-full space-y-2 border-t border-border/50 bg-transparent px-4 py-4">
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
                  'flex w-full items-center rounded-2xl text-sm transition app-nav-inactive',
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
          </div>
        </aside>

        <div className="flex min-w-0 flex-1 flex-col gap-4">
          <header className="flex items-center justify-between app-card px-4 py-4 lg:hidden">
            <button
              type="button"
              onClick={toggleSidebar}
              title="Alternar sidebar"
              className="inline-flex h-11 w-11 items-center justify-center rounded-2xl border border-border/50 bg-surface/50 text-muted transition hover:border-border hover:bg-surface hover:text-foreground"
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

        {currentUser?.role === 'admin' && isDashboardRoute ? (
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
  label: ReactNode;
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
          'app-nav-active',
          collapsedDesktop ? 'justify-center px-0 py-3' : 'gap-3 px-4 py-3',
        ),
      }}
      inactiveProps={{
        className: cn(
          'app-nav-inactive',
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
      aria-hidden={!expanded ? true : undefined}
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
      <div className="flex h-11 w-11 items-center justify-center overflow-hidden rounded-2xl border border-border bg-surface">
        <img
          src={logoUrl}
          alt={`Logo de ${companyName}`}
          className="h-full w-full object-fill"
        />
      </div>
    );
  }

  return (
    <div className="flex h-11 w-11 items-center justify-center rounded-2xl bg-primary font-display text-sm uppercase tracking-[0.18em] text-stone-900 font-bold">
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
          'overflow-hidden rounded-2xl border border-border/50 bg-surface/50',
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
        'flex items-center justify-center rounded-2xl bg-primary font-display uppercase tracking-[0.18em] text-stone-900 font-bold',
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
        <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-surface/50 text-primary shadow-sm border border-border/50">
          <Sparkles className="h-5 w-5" />
        </div>
      </div>
    );
  }

  const title = isSpecialTier
    ? 'Plano consolidado'
    : `Upgrade para ${suggestedPlan}`;

  return (
    <div className="mt-4 rounded-[2rem] app-panel p-6">
      <div className="flex flex-col items-center text-center">
        <h4 className="font-display text-lg text-foreground">{title}</h4>
        <p className="mt-2 text-xs leading-relaxed text-muted px-2">
          {isSpecialTier
            ? 'Você já possui todos os recursos liberados.'
            : 'Ganhe 1 mês grátis e desbloqueie novos recursos agora.'}
        </p>

        <button
          type="button"
          className={cn(
            'mt-6 inline-flex w-full items-center justify-center rounded-2xl bg-surface/80 border border-border/50 px-4 py-3 text-sm font-bold text-primary transition hover:bg-surface shadow-sm',
            isSpecialTier && 'bg-primary text-stone-900 hover:bg-primary/90',
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
    <div className="flex app-shell items-center justify-center px-6 text-center">
      <div className="max-w-sm app-card px-8 py-10">
        <p className="font-display text-3xl text-foreground">PetControl</p>
        <p className="mt-3 text-sm text-muted">
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
