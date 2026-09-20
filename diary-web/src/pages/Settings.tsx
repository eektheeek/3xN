import { useState } from 'preact/hooks';
import { route } from 'preact-router';
import type { RoutableProps } from 'preact-router';
import { api, ApiError } from '../api/client';
import { clearAuth, getAuthUser } from '../auth/session';
import { ErrorBanner } from '../components/ErrorBanner';
import { closeOfflineDB } from '../sync/db';

export function Settings(_props: RoutableProps) {
  const user = getAuthUser();
  const [error, setError] = useState('');
  const [ok, setOk] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const handlePassword = async (e: Event) => {
    e.preventDefault();
    const form = e.target as HTMLFormElement;
    const data = new FormData(form);
    const currentPassword = String(data.get('currentPassword'));
    const newPassword = String(data.get('newPassword'));
    const confirm = String(data.get('confirm'));
    if (newPassword !== confirm) {
      setError('Новые пароли не совпадают');
      setOk('');
      return;
    }
    setSubmitting(true);
    setError('');
    setOk('');
    try {
      await api.changePassword({ currentPassword, newPassword });
      form.reset();
      setOk('Пароль обновлён');
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Не удалось сменить пароль');
    } finally {
      setSubmitting(false);
    }
  };

  const handleLogout = async () => {
    try {
      await api.logout();
    } catch {
      // still clear local session
    }
    await closeOfflineDB();
    clearAuth();
    route('/login', true);
  };

  return (
    <div class="page page--with-tabs">
      <header class="page-header">
        <a href="/more" class="btn btn-ghost">
          ← Назад
        </a>
        <h1>Аккаунт</h1>
      </header>

      <section class="card" style="margin-bottom:1rem">
        <p class="muted" style="margin:0 0 0.35rem">Email</p>
        <p style="margin:0">{user?.email ?? '—'}</p>
      </section>

      <ErrorBanner message={error} />
      {ok ? <p class="muted">{ok}</p> : null}

      <form class="card form" onSubmit={handlePassword}>
        <h2 class="section-title">Сменить пароль</h2>
        <label class="field">
          <span>Текущий пароль</span>
          <input name="currentPassword" type="password" required autocomplete="current-password" />
        </label>
        <label class="field">
          <span>Новый пароль</span>
          <input name="newPassword" type="password" required autocomplete="new-password" minLength={8} />
        </label>
        <label class="field">
          <span>Ещё раз</span>
          <input name="confirm" type="password" required autocomplete="new-password" minLength={8} />
        </label>
        <button type="submit" class="btn btn-primary" disabled={submitting}>
          {submitting ? 'Сохранение…' : 'Сохранить пароль'}
        </button>
      </form>

      <button type="button" class="btn btn-secondary" style="margin-top:1.25rem;width:100%" onClick={() => void handleLogout()}>
        Выйти
      </button>
    </div>
  );
}
