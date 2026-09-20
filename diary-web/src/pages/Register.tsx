import { useState } from 'preact/hooks';
import { route } from 'preact-router';
import type { RoutableProps } from 'preact-router';
import { api, ApiError } from '../api/client';
import { saveAuth } from '../auth/session';
import { ErrorBanner } from '../components/ErrorBanner';

export function Register(_props: RoutableProps) {
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    const form = e.target as HTMLFormElement;
    const data = new FormData(form);
    const password = String(data.get('password'));
    const confirm = String(data.get('confirm'));
    if (password !== confirm) {
      setError('Пароли не совпадают');
      return;
    }
    setSubmitting(true);
    setError('');
    try {
      const res = await api.register({
        email: String(data.get('email')).trim(),
        password,
      });
      saveAuth(res.token, res.user);
      route('/', true);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Не удалось зарегистрироваться');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div class="page">
      <header class="page-header">
        <h1>Регистрация</h1>
      </header>
      <ErrorBanner message={error} />
      <form class="card form" onSubmit={handleSubmit}>
        <label class="field">
          <span>Email</span>
          <input name="email" type="email" required autocomplete="username" />
        </label>
        <label class="field">
          <span>Пароль</span>
          <input name="password" type="password" required autocomplete="new-password" minLength={8} />
        </label>
        <label class="field">
          <span>Ещё раз</span>
          <input name="confirm" type="password" required autocomplete="new-password" minLength={8} />
        </label>
        <button type="submit" class="btn btn-primary" disabled={submitting}>
          {submitting ? 'Создание…' : 'Создать аккаунт'}
        </button>
      </form>
      <p class="muted" style="margin-top:1rem">
        Уже есть аккаунт? <a href="/login">Войти</a>
      </p>
    </div>
  );
}
