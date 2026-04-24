import {
  Activity,
  CalendarDays,
  ChevronLeft,
  ChevronRight,
  MessageSquareText,
  ShieldCheck,
} from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';

import { useInternalChatSocket } from '@/hooks/use-internal-chat-socket';
import {
  useAdminSystemChatMessagesQuery,
  useCompanyUsersQuery,
  useCreateAdminSystemChatMessageMutation,
  useCurrentCompanyQuery,
  useCurrentUserQuery,
} from '@/lib/api/domain.queries';

export function AdminSupportChatAside({
  className = '',
}: {
  className?: string;
}) {
  const companyQuery = useCurrentCompanyQuery();
  const currentUserQuery = useCurrentUserQuery();
  const companyUsersQuery = useCompanyUsersQuery();

  const company = companyQuery.data;
  const currentUser = currentUserQuery.data;
  const now = new Date();
  const greetingName =
    currentUser?.short_name || currentUser?.full_name || company?.fantasy_name;

  const [selectedSystemContactId, setSelectedSystemContactId] =
    useState('contract-pending');
  const [chatDraft, setChatDraft] = useState('');
  const [userStatus, setUserStatus] = useState('online');
  const [chatExpanded, setChatExpanded] = useState(true);
  const chatMessagesContainerRef = useRef<HTMLDivElement | null>(null);

  const preliminarySystemUsers = useMemo(
    () =>
      (companyUsersQuery.data ?? []).filter(
        (user) =>
          user.role === 'system' && user.user_id !== currentUser?.user_id,
      ),
    [companyUsersQuery.data, currentUser?.user_id],
  );

  const effectiveSystemContactId = preliminarySystemUsers.some(
    (user) => user.user_id === selectedSystemContactId,
  )
    ? selectedSystemContactId
    : (preliminarySystemUsers[0]?.user_id ?? undefined);

  const chatMessagesQuery = useAdminSystemChatMessagesQuery(
    effectiveSystemContactId,
  );
  const sendChatMessageMutation = useCreateAdminSystemChatMessageMutation(
    effectiveSystemContactId,
  );
  const { presenceMap, updatePresenceStatus } = useInternalChatSocket(
    effectiveSystemContactId,
  );

  useEffect(() => {
    const container = chatMessagesContainerRef.current;
    if (!container) {
      return;
    }

    container.scrollTo({
      top: container.scrollHeight,
      behavior: 'smooth',
    });
  }, [chatMessagesQuery.data, effectiveSystemContactId]);

  if (
    companyQuery.isLoading ||
    currentUserQuery.isLoading ||
    companyUsersQuery.isLoading
  ) {
    return (
      <aside
        className={`border-l border-stone-100 ${className} hidden xl:flex xl:w-[24rem] xl:flex-col`}
      >
        <div className="p-6 text-sm text-stone-500">
          Carregando chat do sistema...
        </div>
      </aside>
    );
  }

  if (
    !company ||
    !currentUser ||
    currentUser.role !== 'admin' ||
    !greetingName
  ) {
    return null;
  }

  const chatContacts = buildSystemContactOptions(
    companyUsersQuery.data ?? [],
    currentUser.user_id,
  );
  const normalizedSelectedSystemContactId = chatContacts.some(
    (contact) => contact.id === selectedSystemContactId,
  )
    ? selectedSystemContactId
    : (chatContacts[0]?.id ?? 'contract-pending');
  const selectedSystemContact =
    chatContacts.find(
      (contact) => contact.id === normalizedSelectedSystemContactId,
    ) ?? chatContacts[0];

  const contactPresence = effectiveSystemContactId
    ? presenceMap[effectiveSystemContactId]
    : undefined;
  const isContactOnline = contactPresence?.status === 'online';

  const handleStatusChange = (status: string) => {
    setUserStatus(status);
    updatePresenceStatus(status);
  };

  return (
    <aside
      className={`hidden border-l border-stone-100 xl:flex ${className} ${
        chatExpanded ? 'xl:w-[24rem]' : 'xl:w-[5rem]'
      }`}
    >
      <div className="flex min-h-full w-full flex-col divide-y divide-stone-100">
        <div
          className={`border-b border-stone-100 ${
            chatExpanded
              ? 'flex items-center justify-between px-5 py-5'
              : 'flex justify-center px-3 py-5'
          }`}
        >
          {chatExpanded ? (
            <div className="min-w-0">
              <p className="truncate font-display text-xl text-stone-900">
                Chat interno
              </p>
              <p className="truncate text-xs uppercase tracking-[0.28em] text-stone-400">
                suporte admin
              </p>
            </div>
          ) : null}

          <button
            type="button"
            onClick={() => setChatExpanded((current) => !current)}
            title={chatExpanded ? 'Recolher chat' : 'Expandir chat'}
            className="inline-flex h-11 w-11 items-center justify-center rounded-2xl border border-stone-200 bg-stone-50 text-stone-500 transition hover:border-stone-300 hover:bg-stone-100 hover:text-stone-900"
          >
            {chatExpanded ? (
              <ChevronRight className="h-4 w-4" />
            ) : (
              <ChevronLeft className="h-4 w-4" />
            )}
          </button>
        </div>

        {!chatExpanded ? (
          <div className="flex flex-1 flex-col items-center gap-4 px-3 py-6">
            <div className="flex h-12 w-12 items-center justify-center rounded-2xl border border-stone-100 bg-stone-50 text-sky-600 shadow-sm">
              <ShieldCheck className="h-5 w-5" />
            </div>
            <div className="flex h-12 w-12 items-center justify-center rounded-2xl border border-stone-100 bg-stone-50 text-stone-500 shadow-sm">
              <CalendarDays className="h-5 w-5" />
            </div>
            <div className="flex h-12 w-12 items-center justify-center rounded-2xl border border-stone-100 bg-stone-50 text-emerald-500 shadow-sm">
              <Activity className="h-5 w-5" />
            </div>
          </div>
        ) : null}
        {chatExpanded ? (
          <>
            <div className="p-8 text-center">
              <div className="flex flex-col items-center">
                <div className="relative">
                  <div className="h-24 w-24 rounded-full border-4 border-stone-50 bg-stone-100 p-1 shadow-sm">
                    <img
                      src={
                        currentUser.image_url ||
                        `https://ui-avatars.com/api/?name=${greetingName}&background=0D1117&color=fff`
                      }
                      alt={greetingName}
                      className="h-full w-full rounded-full object-cover"
                    />
                  </div>
                  <StatusPicker
                    currentStatus={userStatus}
                    onStatusChange={handleStatusChange}
                  />
                </div>
                <h4 className="mt-2 font-display text-xl text-stone-950">
                  {greetingName}
                </h4>
                <p className="mb-3 text-sm text-stone-400">
                  Administrador {company.fantasy_name}
                </p>

                <div className="mt-6 grid w-full grid-cols-3 gap-3">
                  <MiniBadge icon={ShieldCheck} label="Admin" />
                  <MiniBadge
                    icon={CalendarDays}
                    label={formatCompactDate(now)}
                  />
                  <MiniBadge icon={Activity} label="Chat ativo" />
                </div>
              </div>
            </div>

            <div className="flex flex-1 flex-col p-6 pt-8">
              <div className="border-b border-stone-100 pb-4">
                <div className="flex items-center justify-between gap-3">
                  <div>
                    <p className="text-xs font-semibold uppercase tracking-[0.28em] text-stone-400">
                      Chat do sistema
                    </p>
                    <h5 className="mt-2 font-display text-lg text-stone-950">
                      Suporte ao administrador
                    </h5>
                  </div>
                </div>
                <p className="mt-3 text-sm leading-6 text-stone-500">
                  Este chat persiste mensagens de textos entre os usuários, com
                  suporte a sincronização em tempo real.
                </p>
              </div>

              <div className="mt-5">
                <label
                  htmlFor="shell-system-contact"
                  className="text-xs font-semibold uppercase tracking-[0.24em] text-stone-400"
                >
                  Lista de usuários
                </label>
                <div className="mt-2 rounded-2xl border border-stone-200 bg-stone-50 px-4 py-3">
                  <select
                    id="shell-system-contact"
                    aria-label="Selecionar usuário system"
                    value={normalizedSelectedSystemContactId}
                    onChange={(event) => {
                      setSelectedSystemContactId(event.target.value);
                      setChatDraft('');
                    }}
                    className="w-full bg-transparent text-sm text-stone-700 outline-none"
                  >
                    {chatContacts.map((contact) => (
                      <option key={contact.id} value={contact.id}>
                        {contact.label}
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              <div className="mt-6 flex items-center gap-3 rounded-[1.8rem] border border-stone-100 bg-stone-50/70 p-4">
                <div className="relative">
                  {selectedSystemContact.imageUrl ? (
                    <img
                      src={selectedSystemContact.imageUrl}
                      alt={selectedSystemContact.name}
                      className="h-12 w-12 rounded-full object-cover"
                    />
                  ) : (
                    <div className="flex h-12 w-12 items-center justify-center rounded-full bg-sky-600 text-sm font-semibold uppercase tracking-[0.16em] text-white">
                      {selectedSystemContact.avatar}
                    </div>
                  )}
                  <div
                    className={`absolute -bottom-0.5 -right-0.5 h-3.5 w-3.5 rounded-full border-2 border-white ${
                      isContactOnline ? 'bg-emerald-500' : 'bg-stone-300'
                    }`}
                  />
                </div>
                <div className="min-w-0">
                  <p className="truncate font-medium text-stone-900">
                    {selectedSystemContact.name}
                  </p>
                  <p className="truncate text-sm text-stone-400">
                    {selectedSystemContact.subtitle}
                  </p>
                </div>
              </div>

              <div
                ref={chatMessagesContainerRef}
                className="mt-6 h-[22rem] space-y-5 overflow-y-auto pr-2"
              >
                {!effectiveSystemContactId ? (
                  <div className="rounded-[1.6rem] border border-dashed border-stone-200 bg-stone-50 px-4 py-6 text-sm leading-6 text-stone-500">
                    Vincule um usuário do tipo <strong>sistema</strong> para
                    iniciar uma conversa persistida com o administrador.
                  </div>
                ) : chatMessagesQuery.isLoading ? (
                  <div className="rounded-[1.6rem] border border-stone-100 bg-stone-50 px-4 py-6 text-sm text-stone-500">
                    Carregando histórico da conversa...
                  </div>
                ) : chatMessagesQuery.isError ? (
                  <div className="rounded-[1.6rem] border border-rose-100 bg-rose-50 px-4 py-6 text-sm text-rose-600">
                    Não foi possível carregar o histórico persistido desta
                    conversa.
                  </div>
                ) : (chatMessagesQuery.data?.length ?? 0) === 0 ? (
                  <div className="rounded-[1.6rem] border border-dashed border-stone-200 bg-stone-50 px-4 py-6 text-sm leading-6 text-stone-500">
                    Ainda não existem mensagens persistidas entre este admin e o
                    usuário selecionado.
                  </div>
                ) : (
                  chatMessagesQuery.data?.map((message) => {
                    const isOwnMessage =
                      message.sender_user_id === currentUser.user_id;

                    return (
                      <div
                        key={message.id}
                        className={`flex ${isOwnMessage ? 'justify-end' : 'justify-start'}`}
                      >
                        <div
                          className={`max-w-[88%] rounded-[1.6rem] px-4 py-3 text-sm leading-6 ${
                            isOwnMessage
                              ? 'bg-sky-500 text-white'
                              : 'border border-stone-100 bg-stone-50 text-stone-600'
                          }`}
                        >
                          <p
                            className={`text-[11px] font-semibold uppercase tracking-[0.18em] ${
                              isOwnMessage ? 'text-white/70' : 'text-stone-400'
                            }`}
                          >
                            {message.sender_name}
                          </p>
                          <p className="mt-2 whitespace-pre-wrap">
                            {message.body}
                          </p>
                          <p
                            className={`mt-2 text-[11px] ${
                              isOwnMessage ? 'text-white/70' : 'text-stone-400'
                            }`}
                          >
                            {formatChatTimestamp(message.created_at)}
                          </p>
                        </div>
                      </div>
                    );
                  })
                )}
              </div>

              <form
                className="mt-6 rounded-[1.6rem] border border-stone-200 bg-stone-50 px-4 py-4"
                onSubmit={(event) => {
                  event.preventDefault();
                  const message = chatDraft.trim();
                  if (
                    !effectiveSystemContactId ||
                    !message ||
                    sendChatMessageMutation.isPending
                  ) {
                    return;
                  }

                  sendChatMessageMutation.mutate(
                    { message },
                    {
                      onSuccess: () => {
                        setChatDraft('');
                      },
                    },
                  );
                }}
              >
                <div className="flex items-center gap-3">
                  <MessageSquareText className="h-4 w-4 text-stone-500" />
                  <input
                    id="shell-chat-message"
                    name="message"
                    type="text"
                    autoComplete="off"
                    aria-label="Escrever mensagem para usuário system"
                    value={chatDraft}
                    onChange={(event) => setChatDraft(event.target.value)}
                    placeholder={
                      effectiveSystemContactId
                        ? 'Escreva uma mensagem...'
                        : 'Selecione um usuário system para conversar'
                    }
                    disabled={
                      !effectiveSystemContactId ||
                      sendChatMessageMutation.isPending
                    }
                    className="w-full bg-transparent text-sm text-stone-700 outline-none placeholder:text-stone-400 disabled:cursor-not-allowed"
                  />
                  <button
                    type="submit"
                    disabled={
                      !effectiveSystemContactId ||
                      !chatDraft.trim() ||
                      sendChatMessageMutation.isPending
                    }
                    className="inline-flex items-center justify-center rounded-xl bg-sky-600 px-3 py-2 text-xs font-semibold uppercase tracking-[0.18em] text-white transition hover:bg-sky-700 disabled:cursor-not-allowed disabled:bg-stone-300"
                  >
                    {sendChatMessageMutation.isPending ? 'Enviando' : 'Enviar'}
                  </button>
                </div>
                {sendChatMessageMutation.isError ? (
                  <p className="mt-3 text-sm text-rose-600">
                    Não foi possível persistir a mensagem desta conversa.
                  </p>
                ) : null}
              </form>
            </div>
          </>
        ) : null}
      </div>
    </aside>
  );
}

function MiniBadge({
  icon: Icon,
  label,
}: {
  icon: typeof CalendarDays;
  label: string;
}) {
  return (
    <div className="rounded-2xl border border-stone-100 bg-stone-50 px-3 py-3 text-center">
      <Icon className="mx-auto h-4 w-4 text-stone-500" />
      <p className="mt-2 text-xs font-medium text-stone-500">{label}</p>
    </div>
  );
}

function formatCompactDate(date: Date) {
  return date.toLocaleDateString('pt-BR', {
    day: '2-digit',
    month: '2-digit',
  });
}

function resolveInitials(value: string) {
  const parts = value.trim().split(/\s+/).filter(Boolean).slice(0, 2);

  if (parts.length === 0) {
    return 'PC';
  }

  return parts.map((part) => part[0]?.toUpperCase() ?? '').join('');
}

function buildSystemContactOptions(
  companyUsers: Array<{
    user_id: string;
    role: string;
    short_name?: string | null;
    full_name?: string | null;
    image_url?: string | null;
  }>,
  currentUserId: string,
) {
  const systemUsers = companyUsers.filter(
    (item) => item.role === 'system' && item.user_id !== currentUserId,
  );

  if (systemUsers.length === 0) {
    return [
      {
        id: 'contract-pending',
        label: 'Nenhum usuário vinculado',
        name: 'Usuários',
        subtitle: 'Vincule um contato a empresa para habilitar o seletor',
        avatar: 'SY',
        imageUrl: null,
      },
    ];
  }

  return systemUsers.map((item) => {
    const name = item.short_name || item.full_name || 'Usuário do sistema';
    return {
      id: item.user_id,
      label: name,
      name,
      subtitle: 'Usuário do sistema',
      avatar: resolveInitials(name),
      imageUrl: item.image_url ?? null,
    };
  });
}

function formatChatTimestamp(value: string) {
  const date = new Date(value);

  return new Intl.DateTimeFormat('pt-BR', {
    day: '2-digit',
    month: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date);
}

interface StatusPickerProps {
  currentStatus: string;
  onStatusChange: (status: string) => void;
}

function StatusPicker({ currentStatus, onStatusChange }: StatusPickerProps) {
  const [isOpen, setIsOpen] = useState(false);
  const statusOptions = [
    { id: 'online', label: 'Online', color: 'bg-emerald-500' },
    { id: 'busy', label: 'Ocupado', color: 'bg-rose-500' },
    { id: 'away', label: 'Ausente', color: 'bg-amber-500' },
  ];

  const currentOption =
    statusOptions.find((o) => o.id === currentStatus) || statusOptions[0];

  return (
    <div className="absolute bottom-1 right-1">
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className={`h-5 w-5 rounded-full border-2 border-white shadow-sm transition-all hover:scale-110 active:scale-95 ${currentOption.color}`}
        title="Alterar status de presença"
      />

      {isOpen && (
        <>
          <div
            className="fixed inset-0 z-40"
            onClick={() => setIsOpen(false)}
            onKeyDown={(e) => e.key === 'Escape' && setIsOpen(false)}
            role="presentation"
          />
          <div className="absolute right-0 top-full z-50 mt-2 w-32 origin-top-right animate-in slide-in-from-top-2 fade-in rounded-2xl border border-stone-100 bg-white p-2 shadow-2xl ring-1 ring-black/5 duration-200">
            <div className="flex flex-col gap-1">
              {statusOptions.map((opt) => (
                <button
                  key={opt.id}
                  type="button"
                  onClick={() => {
                    onStatusChange(opt.id);
                    setIsOpen(false);
                  }}
                  className="flex items-center gap-2 rounded-xl px-3 py-2 text-left text-sm text-stone-700 transition hover:bg-stone-50"
                >
                  <div className={`h-3 w-3 rounded-full ${opt.color}`} />
                  <span>{opt.label}</span>
                </button>
              ))}
            </div>
          </div>
        </>
      )}
    </div>
  );
}
