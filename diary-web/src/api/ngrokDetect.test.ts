import { describe, expect, it } from 'vitest';
import { looksLikeNgrokInterstitial } from './ngrokDetect';

describe('looksLikeNgrokInterstitial', () => {
  it('detects html content-type with ngrok markers', () => {
    const html = '<html><body>You are about to visit ngrok — Visit Site</body></html>';
    expect(looksLikeNgrokInterstitial(html, 'text/html; charset=utf-8')).toBe(true);
  });

  it('ignores normal json', () => {
    expect(looksLikeNgrokInterstitial('{"ok":true}', 'application/json')).toBe(false);
  });

  it('detects err_ngrok in body without html content-type', () => {
    expect(looksLikeNgrokInterstitial('ERR_NGROK_3200', 'text/plain')).toBe(true);
  });
});
