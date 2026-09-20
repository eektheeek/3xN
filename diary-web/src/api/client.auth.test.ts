import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { api, ApiError } from './client';
import { clearAuth, getAuthToken, saveAuth } from '../auth/session';

const mem = new Map<string, string>();

beforeEach(() => {
  mem.clear();
  Object.defineProperty(globalThis, 'localStorage', {
    configurable: true,
    value: {
      getItem: (key: string) => mem.get(key) ?? null,
      setItem: (key: string, value: string) => {
        mem.set(key, value);
      },
      removeItem: (key: string) => {
        mem.delete(key);
      },
      clear: () => mem.clear(),
    },
  });
});

afterEach(() => {
  clearAuth();
  vi.restoreAllMocks();
});

describe('api auth headers', () => {
  it('sends Bearer token when logged in', async () => {
    saveAuth('abc123', { id: 'u1', email: 'a@b.c', createdAt: 'now' });
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify([]), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }),
    );
    vi.stubGlobal('fetch', fetchMock);
    await api.listExercises();
    const headers = fetchMock.mock.calls[0][1].headers as Record<string, string>;
    expect(headers.Authorization).toBe('Bearer abc123');
  });

  it('clears session on 401', async () => {
    saveAuth('abc123', { id: 'u1', email: 'a@b.c', createdAt: 'now' });
    Object.defineProperty(globalThis, 'location', {
      configurable: true,
      value: { pathname: '/', assign: vi.fn() },
    });
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response(JSON.stringify({ error: 'authorization required' }), {
          status: 401,
          headers: { 'Content-Type': 'application/json' },
        }),
      ),
    );
    await expect(api.listExercises()).rejects.toBeInstanceOf(ApiError);
    expect(getAuthToken()).toBeNull();
  });
});
