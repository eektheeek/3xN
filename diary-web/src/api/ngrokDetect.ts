/** Detect free-ngrok browser warning HTML instead of JSON API responses. */
export function looksLikeNgrokInterstitial(body: string, contentType: string | null): boolean {
  const ct = (contentType ?? '').toLowerCase();
  if (ct.includes('text/html')) {
    const lower = body.toLowerCase();
    return (
      lower.includes('ngrok') ||
      lower.includes('visit site') ||
      lower.includes('err_ngrok') ||
      lower.includes('ngrok-free')
    );
  }
  const lower = body.toLowerCase();
  return (
    (lower.includes('ngrok') && lower.includes('visit site')) ||
    lower.includes('err_ngrok_')
  );
}
