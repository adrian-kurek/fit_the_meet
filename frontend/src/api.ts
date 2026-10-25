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

  if (response.status === 204 || response.status === 201) return undefined as T;
  return response.json() as Promise<T>;
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
