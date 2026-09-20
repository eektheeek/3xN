import { describe, expect, it } from 'vitest';
import { activeTabForPath, shouldShowTabBar } from './BottomTabBar';

describe('shouldShowTabBar', () => {
  it('shows on root tab routes', () => {
    expect(shouldShowTabBar('/')).toBe(true);
    expect(shouldShowTabBar('/diary')).toBe(true);
    expect(shouldShowTabBar('/stats')).toBe(true);
    expect(shouldShowTabBar('/more')).toBe(true);
    expect(shouldShowTabBar('/cycles')).toBe(true);
    expect(shouldShowTabBar('/protocols')).toBe(true);
    expect(shouldShowTabBar('/settings')).toBe(true);
  });

  it('hides on deep screens', () => {
    expect(shouldShowTabBar('/diary/abc')).toBe(false);
    expect(shouldShowTabBar('/cycles/abc')).toBe(false);
    expect(shouldShowTabBar('/workouts/abc/start')).toBe(false);
    expect(shouldShowTabBar('/exercises/abc')).toBe(false);
  });
});

describe('activeTabForPath', () => {
  it('maps more-section routes to more', () => {
    expect(activeTabForPath('/more')).toBe('more');
    expect(activeTabForPath('/cycles')).toBe('more');
    expect(activeTabForPath('/protocols')).toBe('more');
    expect(activeTabForPath('/settings')).toBe('more');
  });
});
