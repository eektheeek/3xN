type Listener = () => void;

const listeners = new Set<Listener>();
let blocked = false;

export function isTunnelBlocked(): boolean {
  return blocked;
}

export function onTunnelBlockedChange(listener: Listener): () => void {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

function emit() {
  for (const listener of listeners) listener();
}

export function reportTunnelBlocked(): void {
  if (blocked) return;
  blocked = true;
  emit();
}

export function clearTunnelBlocked(): void {
  if (!blocked) return;
  blocked = false;
  emit();
}

/**
 * Full document navigation to a non-SW page so free ngrok can show Visit Site
 * in the PWA cookie jar (not Safari).
 */
export function openTunnelUnlock(): void {
  window.location.assign('/tunnel-unlock.html');
}
