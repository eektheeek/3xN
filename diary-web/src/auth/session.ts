export interface AuthUser {
  id: string;
  email: string;
  createdAt: string;
}

const TOKEN_KEY = 'authToken';
const USER_KEY = 'authUser';

function storage(): Storage | null {
  try {
    if (typeof localStorage === 'undefined') return null;
    return localStorage;
  } catch {
    return null;
  }
}

export function getAuthToken(): string | null {
  return storage()?.getItem(TOKEN_KEY) ?? null;
}

export function getAuthUser(): AuthUser | null {
  const raw = storage()?.getItem(USER_KEY);
  if (!raw) return null;
  try {
    const parsed = JSON.parse(raw) as AuthUser;
    if (!parsed?.id || !parsed.email) return null;
    return parsed;
  } catch {
    return null;
  }
}

export function isLoggedIn(): boolean {
  return Boolean(getAuthToken() && getAuthUser()?.id);
}

export function saveAuth(token: string, user: AuthUser): void {
  storage()?.setItem(TOKEN_KEY, token);
  storage()?.setItem(USER_KEY, JSON.stringify(user));
}

export function clearAuth(): void {
  storage()?.removeItem(TOKEN_KEY);
  storage()?.removeItem(USER_KEY);
}

export function isPublicPath(path: string): boolean {
  const p = path.split('?')[0];
  return p === '/login' || p === '/register';
}
