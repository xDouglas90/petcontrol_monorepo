import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type {
  CreateClientInput,
  CreatePetInput,
  CreateScheduleInput,
  CreateServiceInput,
  ListQueryParams,
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
  clients: (params?: ListQueryParams) => ['domain', 'clients', params ?? {}] as const,
  pets: (params?: ListQueryParams) => ['domain', 'pets', params ?? {}] as const,
  services: (params?: ListQueryParams) => ['domain', 'services', params ?? {}] as const,
  schedules: (params?: ListQueryParams) => ['domain', 'schedules', params ?? {}] as const,
  allClients: () => ['domain', 'clients'] as const,
  allPets: () => ['domain', 'pets'] as const,
  allServices: () => ['domain', 'services'] as const,
  allSchedules: () => ['domain', 'schedules'] as const,
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

export function useSchedulesQuery(params?: ListQueryParams) {
  const session = useAuthStore(selectSession);

  return useQuery({
    queryKey: domainQueryKeys.schedules(params),
    enabled: Boolean(session?.accessToken),
    queryFn: async () => {
      if (!session?.accessToken) {
        throw new Error('Sessão não disponível');
      }
      return listSchedules(session.accessToken, params);
    },
  });
}

export function useClientsQuery(params?: ListQueryParams) {
  const session = useAuthStore(selectSession);

  return useQuery({
    queryKey: domainQueryKeys.clients(params),
    enabled: Boolean(session?.accessToken),
    queryFn: async () => {
      if (!session?.accessToken) {
        throw new Error('Sessão não disponível');
      }
      return listClients(session.accessToken, params);
    },
  });
}

export function usePetsQuery(params?: ListQueryParams) {
  const session = useAuthStore(selectSession);

  return useQuery({
    queryKey: domainQueryKeys.pets(params),
    enabled: Boolean(session?.accessToken),
    queryFn: async () => {
      if (!session?.accessToken) {
        throw new Error('Sessão não disponível');
      }
      return listPets(session.accessToken, params);
    },
  });
}

export function useServicesQuery(params?: ListQueryParams) {
  const session = useAuthStore(selectSession);

  return useQuery({
    queryKey: domainQueryKeys.services(params),
    enabled: Boolean(session?.accessToken),
    queryFn: async () => {
      if (!session?.accessToken) {
        throw new Error('Sessão não disponível');
      }
      return listServices(session.accessToken, params);
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
        queryKey: domainQueryKeys.allSchedules(),
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
        queryKey: domainQueryKeys.allClients(),
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
        queryKey: domainQueryKeys.allClients(),
      });
      await queryClient.invalidateQueries({ queryKey: domainQueryKeys.allPets() });
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
        queryKey: domainQueryKeys.allClients(),
      });
      await queryClient.invalidateQueries({ queryKey: domainQueryKeys.allPets() });
      await queryClient.invalidateQueries({
        queryKey: domainQueryKeys.allSchedules(),
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
      await queryClient.invalidateQueries({ queryKey: domainQueryKeys.allPets() });
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
      await queryClient.invalidateQueries({ queryKey: domainQueryKeys.allPets() });
      await queryClient.invalidateQueries({
        queryKey: domainQueryKeys.allSchedules(),
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
      await queryClient.invalidateQueries({ queryKey: domainQueryKeys.allPets() });
      await queryClient.invalidateQueries({
        queryKey: domainQueryKeys.allSchedules(),
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
        queryKey: domainQueryKeys.allServices(),
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
        queryKey: domainQueryKeys.allServices(),
      });
      await queryClient.invalidateQueries({
        queryKey: domainQueryKeys.allSchedules(),
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
        queryKey: domainQueryKeys.allServices(),
      });
      await queryClient.invalidateQueries({
        queryKey: domainQueryKeys.allSchedules(),
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
        queryKey: domainQueryKeys.allSchedules(),
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
        queryKey: domainQueryKeys.allSchedules(),
      });
    },
  });
}
