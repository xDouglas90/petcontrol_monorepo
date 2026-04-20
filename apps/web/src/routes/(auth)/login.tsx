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
    <main className="min-h-screen bg-stone-50 text-stone-950">
      <div className="mx-auto grid min-h-screen max-w-7xl gap-8 px-4 py-6 lg:grid-cols-[1.1fr_0.9fr] lg:px-8">
        <section className="relative flex flex-col justify-between overflow-hidden rounded-[2.5rem] border border-stone-200 bg-white p-8 shadow-premium lg:p-12">
          <div className="absolute inset-0 bg-[radial-gradient(circle_at_top_right,rgba(2,132,199,0.08),transparent_40%),radial-gradient(circle_at_bottom_left,rgba(16,185,129,0.05),transparent_35%)]" />

          <div className="relative space-y-8">
            <div
              className={cn(
                'inline-flex items-center gap-2 rounded-full border px-4 py-2 text-sm font-medium transition-all duration-300',
                isHealthLoading
                  ? 'border-stone-100 bg-stone-50 text-stone-400'
                  : isHealthError
                    ? 'border-red-100 bg-red-50 text-red-600'
                    : authMode === 'mock'
                      ? 'border-amber-100 bg-amber-50 text-amber-600'
                      : 'border-emerald-100 bg-emerald-50 text-emerald-600',
              )}
            >
              {isHealthLoading ? (
                <Loader2 className="h-4 w-4 animate-spin text-stone-400" />
              ) : isHealthError ? (
                <AlertCircle className="h-4 w-4 text-red-500" />
              ) : authMode === 'mock' ? (
                <Sparkles className="h-4 w-4 text-amber-500" />
              ) : (
                <CheckCircle2 className="h-4 w-4 text-emerald-500" />
              )}
              {isHealthLoading
                ? 'Verificando conexão...'
                : isHealthError
                  ? 'API indisponível'
                  : authMode === 'mock'
                    ? 'Modo experimental ativo'
                    : 'Conexão segura'}
            </div>

            <div className="max-w-2xl space-y-6">
              <div>
                <p className="font-display text-sm uppercase tracking-[0.35em] text-sky-600/80">
                  GroomingFlow
                </p>
                <h1 className="mt-4 font-display text-5xl leading-[1.1] text-stone-900 sm:text-6xl">
                  Gerencie sua PetShop com{' '}
                  <span className="text-sky-600">precisão</span> e elegância.
                </h1>
              </div>
              <p className="max-w-xl text-lg leading-relaxed text-stone-500">
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
          <div className="w-full max-w-lg rounded-[2.5rem] border border-stone-200 bg-white p-8 shadow-premium sm:p-10">
            <div className="mb-10 space-y-2 text-center lg:text-left">
              <p className="text-sm font-semibold uppercase tracking-[0.3em] text-sky-600/80">
                Bem-vindo
              </p>
              <h2 className="font-display text-4xl text-stone-900">
                Acessar Painel
              </h2>
              <p className="text-stone-500">
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
                <div className="rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">
                  {mutation.error.message}
                </div>
              ) : null}

              <button
                type="submit"
                disabled={mutation.isPending}
                className={cn(
                  'group inline-flex w-full items-center justify-center gap-2 rounded-2xl bg-sky-600 px-5 py-4 text-sm font-bold text-white transition-all hover:bg-sky-700 shadow-md hover:shadow-lg active:scale-[0.98]',
                  'disabled:cursor-not-allowed disabled:opacity-70',
                )}
              >
                {mutation.isPending ? 'Autenticando...' : 'Entrar no sistema'}
                <ArrowRight className="h-4 w-4 transition-transform group-hover:translate-x-1" />
              </button>
            </form>

            <div className="mt-10 grid gap-3 rounded-[1.5rem] border border-stone-100 bg-stone-50 p-5 text-sm">
              <div className="flex items-center justify-between gap-4">
                <span className="font-medium text-stone-500">Servidor</span>
                <code className="rounded-full bg-white px-3 py-1 text-xs font-semibold text-sky-600 shadow-sm">
                  {import.meta.env.VITE_API_URL ? 'Cloud API' : 'Localhost'}
                </code>
              </div>
              <div className="flex items-center justify-between gap-4">
                <span className="font-medium text-stone-500">Ambiente</span>
                <code className="rounded-full bg-white px-3 py-1 text-xs font-semibold text-sky-600 shadow-sm uppercase">
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
    <article className="rounded-2xl border border-stone-100 bg-white p-5 shadow-sm transition-all hover:shadow-md">
      <div className="mb-4 flex h-10 w-10 items-center justify-center rounded-xl bg-sky-50 text-sky-600">
        <Icon className="h-5 w-5" />
      </div>
      <h3 className="font-display text-lg text-stone-900">{title}</h3>
      <p className="mt-1 text-sm leading-relaxed text-stone-500">
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
      <span className="text-sm font-semibold text-stone-700">{label}</span>
      <input
        {...inputProps}
        type={type}
        placeholder={placeholder}
        className={cn(
          'w-full rounded-2xl border border-stone-200 bg-stone-50 px-4 py-4 text-stone-900 outline-none transition placeholder:text-stone-400',
          'focus:border-sky-500/50 focus:ring-4 focus:ring-sky-500/10 focus:bg-white',
        )}
      />
      {error ? (
        <span className="text-xs font-medium text-rose-500">{error}</span>
      ) : null}
    </label>
  );
}
