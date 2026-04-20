import { useEffect, useRef, useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { selectSession, useAuthStore } from '@/lib/auth/auth.store';
import { domainQueryKeys } from '@/lib/api/domain.queries';
import { normalizeUrl } from '@petcontrol/shared-utils';
import { API_PATHS, INTERNAL_CHAT_SOCKET } from '@petcontrol/shared-constants';

export interface PresenceInfo {
  user_id: string;
  status: 'online' | 'offline' | 'busy' | 'away' | string;
  last_changed_at: string;
}

export function useInternalChatSocket(counterpartUserId?: string) {
  const session = useAuthStore(selectSession);
  const queryClient = useQueryClient();
  const socketRef = useRef<WebSocket | null>(null);
  const intentionalCloseRef = useRef(false);
  const [isConnected, setIsConnected] = useState(false);
  const [presenceMap, setPresenceMap] = useState<Record<string, PresenceInfo>>({});

  const updatePresenceStatus = (status: string) => {
    if (socketRef.current?.readyState === WebSocket.OPEN) {
      socketRef.current.send(
        JSON.stringify({
          type: 'chat.presence.update',
          status,
        }),
      );
    }
  };

  useEffect(() => {
    if (!session?.accessToken || !counterpartUserId) {
      intentionalCloseRef.current = true;
      return;
    }

    intentionalCloseRef.current = false;
    const apiUrl = normalizeUrl(
      import.meta.env.VITE_API_URL ?? 'http://localhost:8080/api/v1',
    );

    // Convert HTTP to WS and add authentication token as query param.
    const wsUrl = new URL(
      apiUrl.replace(/^http/, 'ws') + API_PATHS.adminSystemChatSocket(counterpartUserId),
    );
    wsUrl.searchParams.set('token', session.accessToken);

    const socket = new WebSocket(
      wsUrl.toString(),
      INTERNAL_CHAT_SOCKET.subprotocol,
    );
    socketRef.current = socket;

    socket.onopen = () => {
      if (intentionalCloseRef.current || socketRef.current !== socket) {
        socket.close();
        return;
      }

      setIsConnected(true);
      console.log('[ChatSocket] Connected');
    };

    socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        handleSocketEvent(data);
      } catch (err) {
        console.error('[ChatSocket] Failed to parse message', err);
      }
    };

    socket.onclose = (event) => {
      if (socketRef.current === socket) {
        socketRef.current = null;
      }

      setIsConnected(false);

      if (intentionalCloseRef.current || event.code === 1000) {
        console.log('[ChatSocket] Closing connection');
        return;
      }

      console.log('[ChatSocket] Disconnected', event.code, event.reason);
    };

    socket.onerror = (err) => {
      if (intentionalCloseRef.current || socketRef.current !== socket) {
        return;
      }

      console.error('[ChatSocket] Error', err);
    };

    function handleSocketEvent(data: Record<string, unknown>) {
      switch (data.type) {
        case 'chat.message.created': {
          const message = data.message as Record<string, unknown>;
          // Update TanStack Query cache for messages
          queryClient.setQueryData<Record<string, unknown>[]>(
            domainQueryKeys.adminSystemChatMessages(counterpartUserId!),
            (old = []) => {
              // Avoid duplicates
              if (old.some((m) => m.id === message.id)) return old;
              return [...old, message];
            },
          );
          break;
        }

        case 'chat.presence.snapshot': {
          const snapshot: Record<string, PresenceInfo> = {};
          const presences = data.presences as PresenceInfo[] | undefined;
          presences?.forEach((p) => {
            snapshot[p.user_id] = p;
          });
          setPresenceMap((prev) => ({ ...prev, ...snapshot }));
          break;
        }

        case 'chat.presence.updated': {
          const presence = data.presence as PresenceInfo;
          setPresenceMap((prev) => ({
            ...prev,
            [presence.user_id]: presence,
          }));
          break;
        }
      }
    }

    return () => {
      intentionalCloseRef.current = true;

      if (socketRef.current === socket) {
        socketRef.current = null;
      }

      setIsConnected(false);
      setPresenceMap({});

      if (
        socket.readyState === WebSocket.OPEN ||
        socket.readyState === WebSocket.CONNECTING
      ) {
        socket.close(1000, 'component cleanup');
      }
    };
  }, [session?.accessToken, counterpartUserId, queryClient]);

  return { isConnected, presenceMap, updatePresenceStatus };
}
