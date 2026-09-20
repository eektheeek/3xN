import type { RoutableProps } from 'preact-router';

export function More(_props: RoutableProps) {
  return (
    <div class="page page--with-tabs">
      <header class="page-header">
        <h1>Ещё</h1>
      </header>

      <ul class="list">
        <li>
          <a href="/settings" class="list-item">
            <span class="list-item__title">Аккаунт</span>
            <span class="list-item__meta">Email, пароль, выход</span>
          </a>
        </li>
        <li>
          <a href="/cycles" class="list-item">
            <span class="list-item__title">Циклы</span>
            <span class="list-item__meta">Дорожки тренировок и шаги</span>
          </a>
        </li>
        <li>
          <a href="/protocols" class="list-item">
            <span class="list-item__title">Табаты</span>
            <span class="list-item__meta">Интервальные протоколы</span>
          </a>
        </li>
      </ul>
    </div>
  );
}
