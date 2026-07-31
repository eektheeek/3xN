import type { JSX } from 'preact';

export type TabId = 'home' | 'diary' | 'stats' | 'more';

const TABS: { id: TabId; href: string; label: string; icon: JSX.Element }[] = [
  {
    id: 'home',
    href: '/',
    label: 'Главная',
    icon: (
      <svg viewBox="0 0 24 24" aria-hidden="true">
        <path
          fill="currentColor"
          d="M12 3.2 4 10v10h5.5v-6h5v6H20V10l-8-6.8Z"
        />
      </svg>
    ),
  },
  {
    id: 'diary',
    href: '/diary',
    label: 'Дневник',
    icon: (
      <svg viewBox="0 0 24 24" aria-hidden="true">
        <path
          fill="currentColor"
          d="M7 3h11a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2Zm0 2v14h11V5H7Zm2 2h7v2H9V7Zm0 4h7v2H9v-2Zm0 4h5v2H9v-2Z"
        />
      </svg>
    ),
  },
  {
    id: 'stats',
    href: '/stats',
    label: 'Статистика',
    icon: (
      <svg viewBox="0 0 24 24" aria-hidden="true">
        <path
          fill="currentColor"
          d="M4 19h16v2H4v-2Zm2-2V9h2v8H6Zm5 0V5h2v12h-2Zm5 0v-5h2v5h-2Z"
        />
      </svg>
    ),
  },
  {
    id: 'more',
    href: '/more',
    label: 'Ещё',
    icon: (
      <svg viewBox="0 0 24 24" aria-hidden="true">
        <path
          fill="currentColor"
          d="M6 10.5a1.5 1.5 0 1 1 0 3 1.5 1.5 0 0 1 0-3Zm6 0a1.5 1.5 0 1 1 0 3 1.5 1.5 0 0 1 0-3Zm6 0a1.5 1.5 0 1 1 0 3 1.5 1.5 0 0 1 0-3Z"
        />
      </svg>
    ),
  },
];

/** Paths where the bottom tab bar is shown. */
export function shouldShowTabBar(path: string): boolean {
  const p = path.split('?')[0];
  return (
    p === '/' ||
    p === '/diary' ||
    p === '/stats' ||
    p === '/more' ||
    p === '/cycles' ||
    p === '/protocols'
  );
}

export function activeTabForPath(path: string): TabId {
  const p = path.split('?')[0];
  if (p === '/diary') return 'diary';
  if (p === '/stats') return 'stats';
  if (p === '/more' || p === '/cycles' || p === '/protocols') return 'more';
  return 'home';
}

interface BottomTabBarProps {
  path: string;
}

export function BottomTabBar({ path }: BottomTabBarProps) {
  const active = activeTabForPath(path);

  return (
    <nav class="tab-bar" aria-label="Основная навигация">
      {TABS.map((tab) => (
        <a
          key={tab.id}
          href={tab.href}
          class={`tab-bar__item${active === tab.id ? ' tab-bar__item--active' : ''}`}
          aria-current={active === tab.id ? 'page' : undefined}
        >
          <span class="tab-bar__icon">{tab.icon}</span>
          <span class="tab-bar__label">{tab.label}</span>
        </a>
      ))}
    </nav>
  );
}
