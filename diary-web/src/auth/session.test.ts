import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { clearAuth, getAuthToken, isLoggedIn, isPublicPath, saveAuth } from './session';

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
});

describe('auth session', () => {
  it('persists token and user', () => {
    saveAuth('tok', { id: 'u1', email: 'a@b.c', createdAt: 'now' });
    expect(getAuthToken()).toBe('tok');
    expect(isLoggedIn()).toBe(true);
  });

  it('clears session', () => {
    saveAuth('tok', { id: 'u1', email: 'a@b.c', createdAt: 'now' });
    clearAuth();
    expect(isLoggedIn()).toBe(false);
    expect(getAuthToken()).toBeNull();
  });

  it('treats login/register as public', () => {
    expect(isPublicPath('/login')).toBe(true);
    expect(isPublicPath('/register')).toBe(true);
    expect(isPublicPath('/')).toBe(false);
  });
});
