import { useState } from 'preact/hooks';
import { route } from 'preact-router';
import type { RoutableProps } from 'preact-router';
import { api, ApiError } from '../api/client';
import { saveAuth } from '../auth/session';
import { ErrorBanner } from '../components/ErrorBanner';

export function Login(_props: RoutableProps) {
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    const form = e.target as HTMLFormElement;
    const data = new FormData(form);
    setSubmitting(true);
    setError('');
    try {
      const res = await api.login({
        email: String(data.get('email')).trim(),
        password: String(data.get('password')),
      });
      saveAuth(res.token, res.user);
      route('/', true);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Не удалось войти');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div class="page">
      <header class="page-header">
        <h1>Вход</h1>
      </header>
      <ErrorBanner message={error} />
      <form class="card form" onSubmit={handleSubmit}>
        <label class="field">
          <span>Email</span>
          <input name="email" type="email" required autocomplete="username" />
        </label>
        <label class="field">
          <span>Пароль</span>
          <input name="password" type="password" required autocomplete="current-password" minLength={8} />
        </label>
        <button type="submit" class="btn btn-primary" disabled={submitting}>
          {submitting ? 'Вход…' : 'Войти'}
        </button>
      </form>
      <p class="muted" style="margin-top:1rem">
        Нет аккаунта? <a href="/register">Регистрация</a>
      </p>
    </div>
  );
}
