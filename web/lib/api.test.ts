import axios, {
  AxiosHeaders,
  type AxiosAdapter,
  type AxiosInstance,
  type InternalAxiosRequestConfig,
} from 'axios';
import { describe, expect, it, vi } from 'vitest';
import { API_BASE_URL, apiUrl, installSessionHandling, mediaUrl } from './api';

type Reply = { status: number };

// Builds an axios instance whose adapter never touches the network. It records
// the config of each request and answers with the next scripted reply.
function newClient(replies: Reply[] = [{ status: 200 }]) {
  const seen: InternalAxiosRequestConfig[] = [];
  let call = 0;
  const adapter: AxiosAdapter = async (config) => {
    seen.push(config);
    const reply = replies[Math.min(call++, replies.length - 1)];
    const response = {
      data: {},
      status: reply.status,
      statusText: '',
      headers: new AxiosHeaders(),
      config,
    };
    if (reply.status >= 400) {
      throw new axios.AxiosError(
        'request failed',
        'ERR_BAD_REQUEST',
        config,
        null,
        response,
      );
    }
    return response;
  };
  const client: AxiosInstance = axios.create({ adapter });
  return { client, seen };
}

function authHeader(config: InternalAxiosRequestConfig): unknown {
  return config.headers.get('Authorization');
}

describe('installSessionHandling', () => {
  it('attaches the stored token to API requests', async () => {
    const { client, seen } = newClient();
    installSessionHandling(client, {
      getToken: () => 'abc',
      onSessionExpired: vi.fn(),
    });

    await client.get(apiUrl('posts'));

    expect(authHeader(seen[0])).toBe('Bearer abc');
  });

  it('sends no Authorization header when there is no token', async () => {
    const { client, seen } = newClient();
    installSessionHandling(client, {
      getToken: () => null,
      onSessionExpired: vi.fn(),
    });

    await client.get(apiUrl('posts'));

    expect(authHeader(seen[0])).toBeUndefined();
  });

  it('does not attach the token to login and register requests', async () => {
    const { client, seen } = newClient();
    installSessionHandling(client, {
      getToken: () => 'abc',
      onSessionExpired: vi.fn(),
    });

    await client.post(apiUrl('auth/login'), {});
    await client.post(apiUrl('auth/register'), {});

    expect(authHeader(seen[0])).toBeUndefined();
    expect(authHeader(seen[1])).toBeUndefined();
  });

  it('does not attach the token to requests for other hosts', async () => {
    const { client, seen } = newClient();
    installSessionHandling(client, {
      getToken: () => 'abc',
      onSessionExpired: vi.fn(),
    });

    await client.get('https://example.com/api/posts');

    expect(authHeader(seen[0])).toBeUndefined();
  });

  it('keeps a token the caller already set', async () => {
    const { client, seen } = newClient();
    installSessionHandling(client, {
      getToken: () => 'stored',
      onSessionExpired: vi.fn(),
    });

    await client.get(apiUrl('posts'), {
      headers: { Authorization: 'Bearer explicit' },
    });

    expect(authHeader(seen[0])).toBe('Bearer explicit');
  });

  it('replaces a placeholder token such as "Bearer null"', async () => {
    const { client, seen } = newClient();
    installSessionHandling(client, {
      getToken: () => 'stored',
      onSessionExpired: vi.fn(),
    });

    await client.get(apiUrl('posts'), {
      headers: { Authorization: 'Bearer null' },
    });

    expect(authHeader(seen[0])).toBe('Bearer stored');
  });

  it('signs the user out when the API rejects a token that was sent', async () => {
    const { client } = newClient([{ status: 401 }]);
    const onSessionExpired = vi.fn();
    installSessionHandling(client, { getToken: () => 'abc', onSessionExpired });

    await expect(client.get(apiUrl('posts'))).rejects.toThrow();

    expect(onSessionExpired).toHaveBeenCalledTimes(1);
  });

  it('refreshes an expired access token and retries the failed request once', async () => {
    const { client, seen } = newClient([{ status: 401 }, { status: 200 }]);
    const refreshAccessToken = vi.fn().mockResolvedValue('fresh-token');
    const setToken = vi.fn();
    const onSessionExpired = vi.fn();
    installSessionHandling(client, {
      getToken: () => 'old-token',
      onSessionExpired,
      refreshAccessToken,
      setToken,
    });

    await client.get(apiUrl('posts'));

    expect(refreshAccessToken).toHaveBeenCalledTimes(1);
    expect(setToken).toHaveBeenCalledWith('fresh-token');
    expect(authHeader(seen[1])).toBe('Bearer fresh-token');
    expect(onSessionExpired).not.toHaveBeenCalled();
  });

  it('signs the user out when refreshing an expired access token fails', async () => {
    const { client } = newClient([{ status: 401 }]);
    const onSessionExpired = vi.fn();
    installSessionHandling(client, {
      getToken: () => 'old-token',
      onSessionExpired,
      refreshAccessToken: vi.fn().mockRejectedValue(new Error('refresh failed')),
    });

    await expect(client.get(apiUrl('posts'))).rejects.toThrow();

    expect(onSessionExpired).toHaveBeenCalledTimes(1);
  });

  it('does not sign out on 401 when no token was sent', async () => {
    const { client } = newClient([{ status: 401 }]);
    const onSessionExpired = vi.fn();
    installSessionHandling(client, { getToken: () => null, onSessionExpired });

    await expect(client.get(apiUrl('posts'))).rejects.toThrow();

    expect(onSessionExpired).not.toHaveBeenCalled();
  });

  it('does not sign out when login itself returns 401', async () => {
    const { client } = newClient([{ status: 401 }]);
    const onSessionExpired = vi.fn();
    installSessionHandling(client, { getToken: () => 'abc', onSessionExpired });

    await expect(client.post(apiUrl('auth/login'), {})).rejects.toThrow();

    expect(onSessionExpired).not.toHaveBeenCalled();
  });

  it('does not sign out on errors other than 401', async () => {
    const { client } = newClient([{ status: 403 }, { status: 500 }]);
    const onSessionExpired = vi.fn();
    installSessionHandling(client, { getToken: () => 'abc', onSessionExpired });

    await expect(client.get(apiUrl('audit-logs'))).rejects.toThrow();
    await expect(client.get(apiUrl('posts'))).rejects.toThrow();

    expect(onSessionExpired).not.toHaveBeenCalled();
  });

  it('stops handling requests after the returned cleanup runs', async () => {
    const { client, seen } = newClient();
    const remove = installSessionHandling(client, {
      getToken: () => 'abc',
      onSessionExpired: vi.fn(),
    });

    remove();
    await client.get(apiUrl('posts'));

    expect(authHeader(seen[0])).toBeUndefined();
  });

  it('only touches the instance it was installed on', async () => {
    const handled = newClient();
    const other = newClient();
    installSessionHandling(handled.client, {
      getToken: () => 'abc',
      onSessionExpired: vi.fn(),
    });

    await other.client.get(apiUrl('posts'));

    expect(authHeader(other.seen[0])).toBeUndefined();
  });
});

describe('apiUrl and mediaUrl', () => {
  it('joins the API base URL and a path with one slash', () => {
    expect(apiUrl('posts')).toBe(`${API_BASE_URL}/posts`);
    expect(apiUrl('/posts')).toBe(`${API_BASE_URL}/posts`);
  });

  it('keeps absolute media URLs', () => {
    expect(mediaUrl('https://cdn.example.com/a.png')).toBe(
      'https://cdn.example.com/a.png',
    );
  });

  it('resolves relative media paths against the API origin', () => {
    const origin = new URL(API_BASE_URL).origin;
    expect(mediaUrl('/uploads/a.png')).toBe(`${origin}/uploads/a.png`);
  });
});
