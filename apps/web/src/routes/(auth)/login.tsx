import { useMutation } from '@tanstack/react-query';
import { Navigate, useNavigate } from '@tanstack/react-router';
import {
  ArrowRight,
  Sparkles,
  ShieldCheck,
  Waves,
  CheckCircle2,
  Loader2,
  AlertCircle,
  type LucideIcon,
} from 'lucide-react';
import { type UseFormRegisterReturn, useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { APP_ROUTES } from '@petcontrol/shared-constants';
import { cn } from '@petcontrol/ui/web';

import {
  login as performLogin,
  getAuthMode,
  ApiError,
} from '@/lib/api/rest-client';
import { useHealthQuery } from '@/lib/api/domain.queries';
import { selectSession, useAuthStore } from '@/lib/auth/auth.store';

const loginSchema = z.object({
  email: z
    .email('Informe um e-mail válido')
    .min(1, 'Informe seu e-mail')
    .trim(),
  password: z.string().min(6, 'A senha deve ter pelo menos 6 caracteres'),
});

type LoginFormValues = z.infer<typeof loginSchema>;

export function LoginPage() {
  const navigate = useNavigate();
  const hydrated = useAuthStore((state) => state.hydrated);
  const session = useAuthStore(selectSession);
  const setSession = useAuthStore((state) => state.setSession);

  const form = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: 'admin@petcontrol.local',
      password: 'password123',
    },
  });

  const mutation = useMutation({
    mutationFn: performLogin,
    onSuccess: (result) => {
      setSession(result);
      void navigate({ to: APP_ROUTES.home });
    },
  });

  const { isLoading: isHealthLoading, isError: isHealthError } =
    useHealthQuery();
  const authMode = getAuthMode();

  if (hydrated && session) {
    return <Navigate to={APP_ROUTES.home} replace />;
  }

  return (
    <main className="min-h-screen bg-surface/50 text-foreground">
      <div className="mx-auto grid min-h-screen max-w-7xl gap-8 px-4 py-6 lg:grid-cols-[1.1fr_0.9fr] lg:px-8">
        <section className="relative flex flex-col justify-between overflow-hidden rounded-[2.5rem] border border-border/50 bg-surface p-8 shadow-premium lg:p-12">
          <div className="absolute inset-0 bg-transparent" />

          <div className="relative space-y-8">
            <div
              className={cn(
                'inline-flex items-center gap-2 rounded-full border px-4 py-2 text-sm font-medium transition-all duration-300',
                isHealthLoading
                  ? 'border-border/50 bg-surface/50 text-muted/70'
                  : isHealthError
                    ? 'border-rose-400/30 bg-rose-500/10 text-rose-300'
                    : authMode === 'mock'
                      ? 'border-amber-400/30 bg-amber-500/10 text-amber-300'
                      : 'border-emerald-400/30 bg-emerald-500/10 text-emerald-300',
              )}
            >
              {isHealthLoading ? (
                <Loader2 className="h-4 w-4 animate-spin text-muted/70" />
              ) : isHealthError ? (
                <AlertCircle className="h-4 w-4 text-rose-400" />
              ) : authMode === 'mock' ? (
                <Sparkles className="h-4 w-4 text-amber-400" />
              ) : (
                <CheckCircle2 className="h-4 w-4 text-emerald-400" />
              )}
              {isHealthLoading
                ? 'Verificando conexão...'
                : isHealthError
                  ? 'Servidor Indisponível'
                  : authMode === 'mock'
                    ? 'Modo experimental ativo'
                    : 'Servidor Online'}
            </div>

            <div className="max-w-2xl space-y-6">
              <div>
                <img
                  aria-label="GroomingFlow Logo"
                  src="https://storage.googleapis.com/petcontrol_bucket/assets/images/logo-full.png"
                  alt="GroomingFlow"
                  className="h-auto w-60"
                />
                <h1 className="mt-4 font-display text-5xl leading-[1.1] text-foreground sm:text-6xl">
                  Gerencie sua PetShop com{' '}
                  <span className="text-sky-400">precisão</span> e elegância.
                </h1>
              </div>
              <p className="max-w-xl text-lg leading-relaxed text-muted">
                Uma plataforma completa para gestão pet: de agendamentos
                complexos a controle financeiro, tudo em um só lugar.
              </p>
            </div>
          </div>

          <div className="relative grid gap-4 sm:grid-cols-3">
            <FeatureCard
              icon={ShieldCheck}
              title="Segurança"
              description="Proteção de dados em nível bancário"
            />
            <FeatureCard
              icon={Waves}
              title="Performance"
              description="Interface ultra rápida e responsiva"
            />
            <FeatureCard
              icon={ArrowRight}
              title="Escalável"
              description="Pronto para crescer com seu negócio"
            />
          </div>
        </section>

        <section className="flex items-center justify-center">
          <div className="w-full max-w-lg rounded-[2.5rem] border border-border/50 bg-surface p-8 shadow-premium sm:p-10">
            <div className="mb-10 space-y-2 text-center lg:text-left">
              <p className="text-sm font-semibold uppercase tracking-[0.3em] text-primary/80">
                Bem-vindo
              </p>
              <h2 className="font-display text-4xl text-foreground">
                Acessar Painel
              </h2>
              <p className="text-muted">
                Entre com suas credenciais para continuar.
              </p>
            </div>

            <form
              className="space-y-6"
              onSubmit={form.handleSubmit((values) => mutation.mutate(values))}
            >
              <Field
                label="E-mail"
                error={form.formState.errors.email?.message}
                inputProps={form.register('email')}
                type="email"
                placeholder="admin@petcontrol.local"
              />

              <Field
                label="Senha"
                error={form.formState.errors.password?.message}
                inputProps={form.register('password')}
                type="password"
                placeholder="••••••••"
              />

              {mutation.error instanceof ApiError ? (
                <div className="rounded-2xl border border-rose-400/30 bg-rose-500/10 px-4 py-3 text-sm text-rose-300">
                  {mutation.error.message}
                </div>
              ) : null}

              <button
                type="submit"
                disabled={mutation.isPending}
                className={cn(
                  'group inline-flex w-full items-center justify-center gap-2 rounded-2xl bg-sky-600 px-5 py-4 text-sm font-bold text-white transition-all hover:bg-sky-500 shadow-md hover:shadow-lg active:scale-[0.98]',
                  'disabled:cursor-not-allowed disabled:opacity-70',
                )}
              >
                {mutation.isPending ? 'Autenticando...' : 'Entrar no sistema'}
                <ArrowRight className="h-4 w-4 transition-transform group-hover:translate-x-1" />
              </button>
            </form>

            <div className="mt-10 grid gap-3 rounded-[1.5rem] border border-border/50 bg-surface/30 p-5 text-sm">
              <div className="flex items-center justify-between gap-4">
                <span className="font-medium text-muted">Servidor</span>
                <code className="rounded-full border border-border/50 bg-surface px-3 py-1 text-xs font-semibold text-primary shadow-sm">
                  {import.meta.env.VITE_API_URL ? 'Cloud API' : 'Localhost'}
                </code>
              </div>
              <div className="flex items-center justify-between gap-4">
                <span className="font-medium text-muted">Ambiente</span>
                <code className="rounded-full border border-border/50 bg-surface px-3 py-1 text-xs font-semibold text-primary shadow-sm uppercase">
                  {getAuthMode()}
                </code>
              </div>
            </div>
          </div>
        </section>
      </div>
    </main>
  );
}

function FeatureCard({
  icon: Icon,
  title,
  description,
}: {
  icon: LucideIcon;
  title: string;
  description: string;
}) {
  return (
    <article className="rounded-2xl border border-border/50 bg-surface/30 p-5 transition-all hover:bg-surface/60 hover:border-border">
      <div className="mb-4 flex h-10 w-10 items-center justify-center rounded-xl border border-border/50 bg-surface text-primary shadow-sm">
        <Icon className="h-5 w-5" />
      </div>
      <h3 className="font-display text-lg text-foreground">{title}</h3>
      <p className="mt-1 text-sm leading-relaxed text-muted">
        {description}
      </p>
    </article>
  );
}

function Field({
  label,
  error,
  inputProps,
  type,
  placeholder,
}: {
  label: string;
  error?: string;
  inputProps: UseFormRegisterReturn;
  type: string;
  placeholder: string;
}) {
  return (
    <label className="block space-y-2">
      <span className="text-sm font-semibold text-foreground">{label}</span>
      <input
        {...inputProps}
        type={type}
        placeholder={placeholder}
        className={cn(
          'w-full rounded-2xl border border-border/50 bg-surface/50 px-4 py-4 text-foreground outline-none transition placeholder:text-muted/70',
          'focus:border-sky-500/50 focus:ring-4 focus:ring-sky-500/10 focus:bg-surface',
        )}
      />
      {error ? (
        <span className="text-xs font-medium text-rose-500">{error}</span>
      ) : null}
    </label>
  );
}
