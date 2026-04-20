import { useMutation, useQueries, useQuery, useQueryClient } from '@tanstack/react-query';
import type {
  CompanyUserDTO,
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
  getScheduleHistory,
  getCurrentCompany,
  getCurrentCompanySystemConfig,
  getCurrentUser,
  listCompanyUsers,
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

const EMPTY_PARAMS: ListQueryParams = {};

export const domainQueryKeys = {
  currentCompany: () => ['domain', 'company', 'current'] as const,
  currentCompanySystemConfig: () =>
    ['domain', 'company', 'system-config', 'current'] as const,
  currentUser: () => ['domain', 'user', 'current'] as const,
  companyUsers: () => ['domain', 'company-users'] as const,
  clients: (params?: ListQueryParams) =>
    ['domain', 'clients', params ?? EMPTY_PARAMS] as const,
  pets: (params?: ListQueryParams) =>
    ['domain', 'pets', params ?? EMPTY_PARAMS] as const,
  services: (params?: ListQueryParams) =>
    ['domain', 'services', params ?? EMPTY_PARAMS] as const,
  schedules: (params?: ListQueryParams) =>
    ['domain', 'schedules', params ?? EMPTY_PARAMS] as const,
  scheduleHistory: (scheduleId: string) =>
    ['domain', 'schedules', scheduleId, 'history'] as const,
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

export function useCurrentCompanySystemConfigQuery() {
  const session = useAuthStore(selectSession);

  return useQuery({
    queryKey: domainQueryKeys.currentCompanySystemConfig(),
    enabled: Boolean(session?.accessToken),
    queryFn: async () => {
      if (!session?.accessToken) {
        throw new Error('Sessão não disponível');
      }
      return getCurrentCompanySystemConfig(session.accessToken);
    },
  });
}

export function useCurrentUserQuery() {
  const session = useAuthStore(selectSession);

  return useQuery({
    queryKey: domainQueryKeys.currentUser(),
    enabled: Boolean(session?.accessToken),
    queryFn: async () => {
      if (!session?.accessToken) {
        throw new Error('Sessão não disponível');
      }
      return getCurrentUser(session.accessToken);
    },
  });
}

export function useCompanyUsersQuery() {
  const session = useAuthStore(selectSession);

  return useQuery<CompanyUserDTO[]>({
    queryKey: domainQueryKeys.companyUsers(),
    enabled: Boolean(session?.accessToken),
    queryFn: async () => {
      if (!session?.accessToken) {
        throw new Error('Sessão não disponível');
      }
      return listCompanyUsers(session.accessToken);
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

export function useScheduleHistoriesQuery(scheduleIds: string[]) {
  const session = useAuthStore(selectSession);

  return useQueries({
    queries: scheduleIds.map((scheduleId) => ({
      queryKey: domainQueryKeys.scheduleHistory(scheduleId),
      enabled: Boolean(session?.accessToken && scheduleId),
      queryFn: async () => {
        if (!session?.accessToken) {
          throw new Error('Sessão não disponível');
        }
        return getScheduleHistory(session.accessToken, scheduleId);
      },
    })),
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
      await queryClient.invalidateQueries({
        queryKey: domainQueryKeys.allPets(),
      });
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
      await queryClient.invalidateQueries({
        queryKey: domainQueryKeys.allPets(),
      });
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
      await queryClient.invalidateQueries({
        queryKey: domainQueryKeys.allPets(),
      });
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
      await queryClient.invalidateQueries({
        queryKey: domainQueryKeys.allPets(),
      });
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
      await queryClient.invalidateQueries({
        queryKey: domainQueryKeys.allPets(),
      });
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
