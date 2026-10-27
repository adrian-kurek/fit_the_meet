export type LoginCredentials = {
  email: string;
  password: string;
};

export type RegistrationDetails = LoginCredentials & {
  username: string;
};

export type EventInput = {
  name: string;
  description: string;
  duration_minutes: number;
};

export type AvailabilityTimeBlock = {
  date: string;
  start_time: string;
  end_time: string;
};

export type PollInput = EventInput & {
  options: Array<{
    start_time: string;
    end_time: string;
  }>;
};

type ErrorResponse = {
  message?: string;
};

const apiBaseUrl = (import.meta.env.VITE_API_BASE_URL || '/api').replace(/\/$/, '');

export class ApiError extends Error {
  status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

async function request<T>(path: string, options: RequestInit = {}, accessToken?: string): Promise<T> {
  const headers = new Headers(options.headers);
  if (options.body) headers.set('Content-Type', 'application/json');
  if (accessToken) headers.set('Authorization', `Bearer ${accessToken}`);

  let response: Response;
  try {
    response = await fetch(`${apiBaseUrl}${path}`, {
      ...options,
      headers,
      credentials: 'include',
    });
  } catch {
    throw new ApiError('Could not reach the API. Check that the Go server is running on port 8080.', 0);
  }

  if (!response.ok) {
    let message = `Request failed with status ${response.status}.`;
    try {
      const error = await response.json() as ErrorResponse;
      if (error.message) message = error.message;
    } catch {
      // The backend may return an empty body for some errors.
    }
    throw new ApiError(message, response.status);
  }

  if (response.status === 204) return undefined as T;

  const responseBody = await response.text();
  if (!responseBody) return undefined as T;
  return JSON.parse(responseBody) as T;
}

export const authApi = {
  async login(credentials: LoginCredentials) {
    return request<{ token: string }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(credentials),
    });
  },

  async register(details: RegistrationDetails) {
    return request<void>('/auth/register', {
      method: 'POST',
      body: JSON.stringify(details),
    });
  },

  async logout(accessToken: string) {
    return request<void>('/auth/logout', { method: 'DELETE' }, accessToken);
  },
};

export const eventApi = {
  async create(input: EventInput, accessToken: string, groupId?: number) {
    const path = groupId ? `/groups/${groupId}/events/` : '/events/';
    return request<void>(path, {
      method: 'POST',
      body: JSON.stringify(input),
    }, accessToken);
  },

  async toggleActive(eventId: number, accessToken: string) {
    return request<void>(`/events/${eventId}`, { method: 'PATCH' }, accessToken);
  },

  async saveAvailability(eventId: number, timeBlocks: AvailabilityTimeBlock[], accessToken: string) {
    return request<void>(`/events/${eventId}`, {
      method: 'PUT',
      body: JSON.stringify({ time_blocks: timeBlocks }),
    }, accessToken);
  },
};

export const pollApi = {
  async create(input: PollInput, accessToken: string, groupId?: number) {
    const path = groupId ? `/groups/${groupId}/polls/` : '/polls/';
    return request<void>(path, {
      method: 'POST',
      body: JSON.stringify(input),
    }, accessToken);
  },

  async submitAnswers(pollId: number, optionIds: number[], accessToken: string) {
    return request<void>(`/polls/${pollId}/answers`, {
      method: 'POST',
      body: JSON.stringify({ answers: optionIds.map((option_id) => ({ option_id })) }),
    }, accessToken);
  },

  async toggleActive(pollId: number, accessToken: string) {
    return request<void>(`/polls/${pollId}`, { method: 'PATCH' }, accessToken);
  },
};

export const groupApi = {
  async create(name: string, accessToken: string) {
    return request<{ group: { id: number; name: string } }>('/groups/', {
      method: 'POST',
      body: JSON.stringify({ name }),
    }, accessToken);
  },

  async join(code: string, accessToken: string) {
    return request<{ group: { id: number; name: string; role: 'owner' | 'member'; member_count: number } }>(
      `/groups/?code=${encodeURIComponent(code)}`,
      { method: 'PATCH' },
      accessToken,
    );
  },

  async generateActivationCode(groupId: number, accessToken: string) {
    return request<{ code: string }>(`/groups/${groupId}`, { method: 'POST' }, accessToken);
  },

  async removeParticipants(groupId: number, userIds: number[], accessToken: string) {
    return request<void>(`/groups/${groupId}/participants`, {
      method: 'DELETE',
      body: JSON.stringify({ userIds }),
    }, accessToken);
  },
};
