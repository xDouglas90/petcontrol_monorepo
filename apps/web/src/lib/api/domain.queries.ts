import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type {
  CreateClientInput,
  CreatePetInput,
  CreateScheduleInput,
  CreateServiceInput,
  UpdateClientInput,
  UpdatePetInput,
  UpdateScheduleInput,
  UpdateServiceInput,
} from '@petcontrol/shared-types';

import {
  createClient,
  createPet,
  createSchedule,
  createService,
  deleteClient,
  deletePet,
  deleteSchedule,
  deleteService,
  getCurrentCompany,
  listClients,
  listPets,
  listSchedules,
  listServices,
  updateClient,
  updatePet,
  updateSchedule,
  updateService,
} from '@/lib/api/rest-client';
import { selectSession, useAuthStore } from '@/lib/auth/auth.store';

export const domainQueryKeys = {
  currentCompany: () => ['domain', 'company', 'current'] as const,
  clients: () => ['domain', 'clients'] as const,
  pets: () => ['domain', 'pets'] as const,
  services: () => ['domain', 'services'] as const,
  schedules: () => ['domain', 'schedules'] as const,
};

export function useCurrentCompanyQuery() {
  const session = useAuthStore(selectSession);

  return useQuery({
    queryKey: domainQueryKeys.currentCompany(),
    enabled: Boolean(session?.accessToken),
    queryFn: async () => {
      if (!session?.accessToken) {
        throw new Error('Sessão não disponível');
      }
      return getCurrentCompany(session.accessToken);
    },
  });
}

export function useSchedulesQuery() {
  const session = useAuthStore(selectSession);

  return useQuery({
    queryKey: domainQueryKeys.schedules(),
    enabled: Boolean(session?.accessToken),
    queryFn: async () => {
      if (!session?.accessToken) {
        throw new Error('Sessão não disponível');
      }
      return listSchedules(session.accessToken);
    },
  });
}

export function useClientsQuery() {
  const session = useAuthStore(selectSession);

  return useQuery({
    queryKey: domainQueryKeys.clients(),
    enabled: Boolean(session?.accessToken),
    queryFn: async () => {
      if (!session?.accessToken) {
        throw new Error('Sessão não disponível');
      }
      return listClients(session.accessToken);
    },
  });
}

export function usePetsQuery() {
  const session = useAuthStore(selectSession);

  return useQuery({
    queryKey: domainQueryKeys.pets(),
    enabled: Boolean(session?.accessToken),
    queryFn: async () => {
      if (!session?.accessToken) {
        throw new Error('Sessão não disponível');
      }
      return listPets(session.accessToken);
    },
  });
}

export function useServicesQuery() {
  const session = useAuthStore(selectSession);

  return useQuery({
    queryKey: domainQueryKeys.services(),
    enabled: Boolean(session?.accessToken),
    queryFn: async () => {
      if (!session?.accessToken) {
        throw new Error('Sessão não disponível');
      }
      return listServices(session.accessToken);
    },
  });
}

export function useCreateScheduleMutation() {
  const queryClient = useQueryClient();
  const session = useAuthStore(selectSession);

  return useMutation({
    mutationFn: async (input: CreateScheduleInput) => {
      if (!session?.accessToken) {
        throw new Error('Sessão não disponível');
      }
      return createSchedule(session.accessToken, input);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: domainQueryKeys.schedules(),
      });
    },
  });
}

export function useCreateClientMutation() {
  const queryClient = useQueryClient();
  const session = useAuthStore(selectSession);

  return useMutation({
    mutationFn: async (input: CreateClientInput) => {
      if (!session?.accessToken) {
        throw new Error('Sessão não disponível');
      }
      return createClient(session.accessToken, input);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: domainQueryKeys.clients(),
      });
    },
  });
}

export function useUpdateClientMutation() {
  const queryClient = useQueryClient();
  const session = useAuthStore(selectSession);

  return useMutation({
    mutationFn: async ({
      clientId,
      input,
    }: {
      clientId: string;
      input: UpdateClientInput;
    }) => {
      if (!session?.accessToken) {
        throw new Error('Sessão não disponível');
      }
      return updateClient(session.accessToken, clientId, input);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: domainQueryKeys.clients(),
      });
      await queryClient.invalidateQueries({ queryKey: domainQueryKeys.pets() });
    },
  });
}

export function useDeleteClientMutation() {
  const queryClient = useQueryClient();
  const session = useAuthStore(selectSession);

  return useMutation({
    mutationFn: async (clientId: string) => {
      if (!session?.accessToken) {
        throw new Error('Sessão não disponível');
      }
      await deleteClient(session.accessToken, clientId);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: domainQueryKeys.clients(),
      });
      await queryClient.invalidateQueries({ queryKey: domainQueryKeys.pets() });
      await queryClient.invalidateQueries({
        queryKey: domainQueryKeys.schedules(),
      });
    },
  });
}

export function useCreatePetMutation() {
  const queryClient = useQueryClient();
  const session = useAuthStore(selectSession);

  return useMutation({
    mutationFn: async (input: CreatePetInput) => {
      if (!session?.accessToken) {
        throw new Error('Sessão não disponível');
      }
      return createPet(session.accessToken, input);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: domainQueryKeys.pets() });
    },
  });
}

export function useUpdatePetMutation() {
  const queryClient = useQueryClient();
  const session = useAuthStore(selectSession);

  return useMutation({
    mutationFn: async ({
      petId,
      input,
    }: {
      petId: string;
      input: UpdatePetInput;
    }) => {
      if (!session?.accessToken) {
        throw new Error('Sessão não disponível');
      }
      return updatePet(session.accessToken, petId, input);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: domainQueryKeys.pets() });
      await queryClient.invalidateQueries({
        queryKey: domainQueryKeys.schedules(),
      });
    },
  });
}

export function useDeletePetMutation() {
  const queryClient = useQueryClient();
  const session = useAuthStore(selectSession);

  return useMutation({
    mutationFn: async (petId: string) => {
      if (!session?.accessToken) {
        throw new Error('Sessão não disponível');
      }
      await deletePet(session.accessToken, petId);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: domainQueryKeys.pets() });
      await queryClient.invalidateQueries({
        queryKey: domainQueryKeys.schedules(),
      });
    },
  });
}

export function useCreateServiceMutation() {
  const queryClient = useQueryClient();
  const session = useAuthStore(selectSession);

  return useMutation({
    mutationFn: async (input: CreateServiceInput) => {
      if (!session?.accessToken) {
        throw new Error('Sessão não disponível');
      }
      return createService(session.accessToken, input);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: domainQueryKeys.services(),
      });
    },
  });
}

export function useUpdateServiceMutation() {
  const queryClient = useQueryClient();
  const session = useAuthStore(selectSession);

  return useMutation({
    mutationFn: async ({
      serviceId,
      input,
    }: {
      serviceId: string;
      input: UpdateServiceInput;
    }) => {
      if (!session?.accessToken) {
        throw new Error('Sessão não disponível');
      }
      return updateService(session.accessToken, serviceId, input);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: domainQueryKeys.services(),
      });
      await queryClient.invalidateQueries({
        queryKey: domainQueryKeys.schedules(),
      });
    },
  });
}

export function useDeleteServiceMutation() {
  const queryClient = useQueryClient();
  const session = useAuthStore(selectSession);

  return useMutation({
    mutationFn: async (serviceId: string) => {
      if (!session?.accessToken) {
        throw new Error('Sessão não disponível');
      }
      await deleteService(session.accessToken, serviceId);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: domainQueryKeys.services(),
      });
      await queryClient.invalidateQueries({
        queryKey: domainQueryKeys.schedules(),
      });
    },
  });
}

export function useUpdateScheduleMutation() {
  const queryClient = useQueryClient();
  const session = useAuthStore(selectSession);

  return useMutation({
    mutationFn: async ({
      scheduleId,
      input,
    }: {
      scheduleId: string;
      input: UpdateScheduleInput;
    }) => {
      if (!session?.accessToken) {
        throw new Error('Sessão não disponível');
      }
      return updateSchedule(session.accessToken, scheduleId, input);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: domainQueryKeys.schedules(),
      });
    },
  });
}

export function useDeleteScheduleMutation() {
  const queryClient = useQueryClient();
  const session = useAuthStore(selectSession);

  return useMutation({
    mutationFn: async (scheduleId: string) => {
      if (!session?.accessToken) {
        throw new Error('Sessão não disponível');
      }
      await deleteSchedule(session.accessToken, scheduleId);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: domainQueryKeys.schedules(),
      });
    },
  });
}
