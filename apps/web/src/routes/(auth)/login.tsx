import { useMutation } from '@tanstack/react-query';
import { Navigate, useNavigate } from '@tanstack/react-router';
import {
  ArrowRight,
  Sparkles,
  ShieldCheck,
  Waves,
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
import { selectSession, useAuthStore } from '@/lib/auth/auth.store';

const loginSchema = z.object({
  email: z
    .string()
    .trim()
    .min(1, 'Informe seu e-mail')
    .email('Informe um e-mail válido'),
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

  if (hydrated && session) {
    return <Navigate to={APP_ROUTES.home} replace />;
  }

  return (
    <main className="min-h-screen bg-hero-radial text-foreground">
      <div className="mx-auto grid min-h-screen max-w-7xl gap-8 px-4 py-6 lg:grid-cols-[1.15fr_0.85fr] lg:px-8">
        <section className="relative overflow-hidden rounded-[2rem] border border-white/10 bg-white/5 p-8 shadow-glow backdrop-blur-xl lg:p-12">
          <div className="absolute inset-0 bg-[radial-gradient(circle_at_top_right,rgba(245,158,11,0.18),transparent_34%),radial-gradient(circle_at_bottom_left,rgba(56,189,248,0.18),transparent_30%)]" />
          <div className="relative flex h-full flex-col justify-between gap-10">
            <div className="space-y-6">
              <div className="inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/10 px-4 py-2 text-sm font-medium text-white/80">
                <Sparkles className="h-4 w-4 text-secondary" />
                {getAuthMode() === 'mock'
                  ? 'Modo mock ativo'
                  : 'Login conectado à API'}
              </div>
              <div className="max-w-2xl space-y-4">
                <p className="font-display text-sm uppercase tracking-[0.35em] text-secondary/80">
                  GroomingFlow
                </p>
                <h1 className="font-display text-5xl leading-tight text-white sm:text-6xl">
                  Uma base operacional para pet shops que precisam de controle
                  real.
                </h1>
                <p className="max-w-xl text-lg leading-8 text-white/72">
                  Multi-tenant, plano por módulo, auth com JWT e um frontend já
                  pronto para evoluir com o produto.
                </p>
              </div>
            </div>

            <div className="grid gap-4 sm:grid-cols-3">
              <FeatureCard
                icon={ShieldCheck}
                title="Autenticação"
                description="JWT, bcrypt e middleware de tenant"
              />
              <FeatureCard
                icon={Waves}
                title="Estado"
                description="TanStack Query para API e Zustand para UI"
              />
              <FeatureCard
                icon={ArrowRight}
                title="Fluxo"
                description="Login direto para o dashboard"
              />
            </div>
          </div>
        </section>

        <section className="flex items-center justify-center">
          <div className="w-full max-w-lg rounded-[2rem] border border-white/10 bg-slate-950/75 p-6 shadow-glow backdrop-blur-xl sm:p-8">
            <div className="mb-8 space-y-2">
              <p className="text-sm font-semibold uppercase tracking-[0.3em] text-secondary/80">
                Acesso
              </p>
              <h2 className="font-display text-3xl text-white">
                Entre no painel
              </h2>
              <p className="text-sm text-slate-300">
                Use a conta criada na API ou o mock controlado via ambiente.
              </p>
            </div>

            <form
              className="space-y-5"
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
                placeholder="Sua senha"
              />

              {mutation.error instanceof ApiError ? (
                <div className="rounded-2xl border border-rose-500/30 bg-rose-500/10 px-4 py-3 text-sm text-rose-100">
                  {mutation.error.message}
                </div>
              ) : null}

              <button
                type="submit"
                disabled={mutation.isPending}
                className={cn(
                  'inline-flex w-full items-center justify-center gap-2 rounded-2xl bg-primary px-5 py-3 text-sm font-semibold text-slate-950 transition hover:brightness-110',
                  'disabled:cursor-not-allowed disabled:opacity-70',
                )}
              >
                {mutation.isPending ? 'Autenticando...' : 'Entrar no painel'}
                <ArrowRight className="h-4 w-4" />
              </button>
            </form>

            <div className="mt-8 grid gap-3 rounded-2xl border border-white/8 bg-white/5 p-4 text-sm text-slate-300">
              <div className="flex items-center justify-between gap-4">
                <span className="font-medium text-white/80">API</span>
                <code className="rounded-full bg-black/30 px-3 py-1 text-xs text-secondary">
                  {import.meta.env.VITE_API_URL ??
                    'http://localhost:8080/api/v1'}
                </code>
              </div>
              <div className="flex items-center justify-between gap-4">
                <span className="font-medium text-white/80">Modo</span>
                <code className="rounded-full bg-black/30 px-3 py-1 text-xs text-secondary">
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
    <article className="rounded-2xl border border-white/10 bg-slate-900/55 p-4 backdrop-blur-md">
      <Icon className="mb-4 h-5 w-5 text-primary" />
      <h3 className="font-display text-lg text-white">{title}</h3>
      <p className="mt-1 text-sm leading-6 text-white/68">{description}</p>
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
      <span className="text-sm font-medium text-slate-200">{label}</span>
      <input
        {...inputProps}
        type={type}
        placeholder={placeholder}
        className={cn(
          'w-full rounded-2xl border border-white/10 bg-white/5 px-4 py-3 text-slate-100 outline-none transition placeholder:text-slate-500',
          'focus:border-primary/50 focus:ring-2 focus:ring-primary/20',
        )}
      />
      {error ? <span className="text-sm text-rose-300">{error}</span> : null}
    </label>
  );
}
