/**
 * AssinaPix Brand Tokens
 *
 * Source of truth para cores, tipografia e assets da marca.
 * Os valores aqui espelham os CSS custom properties em globals.css.
 */

export const colors = {
  primary:      '#E8622A',
  primaryDark:  '#BF4A1A',
  primaryLight: '#FDE8DE',

  accent:       '#16A394',
  accentDark:   '#0D7A6E',
  accentLight:  '#D4F5F0',

  background:   '#FAFAF8',
  surface:      '#FFFFFF',
  border:       '#E8E8E5',
  muted:        '#F4F4F2',
  mutedFg:      '#737370',
  foreground:   '#1A1A18',

  success:      '#22C55E',
  warning:      '#F59E0B',
  danger:       '#EF4444',
} as const;

export const typography = {
  fontHeadline: "'Satoshi', -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif",
  fontBody:     "system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif",
  weights: {
    regular:    '400',
    medium:     '500',
    semibold:   '600',
    bold:       '700',
    black:      '900',
  },
} as const;

export const assets = {
  logo:       '/brand/logo.svg',
  logoDark:   '/brand/logo-dark.svg',
  logoIcon:   '/brand/logo-icon.svg',
} as const;

export const brand = { colors, typography, assets } as const;
