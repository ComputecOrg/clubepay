import { describe, it, expect, beforeEach } from 'vitest';
import { renderHook } from '@testing-library/react';
import React from 'react';
import { useAnalytics } from '@/hooks/useAnalytics';
import { AnalyticsProvider } from '@/lib/analyticsContext';

describe('useAnalytics hook', () => {
  beforeEach(() => {
    delete (window as any).gtag;
    delete (window as any).fbq;
  });

  it('should return analytics instance from context', () => {
    const measurementId = 'G-TEST123';
    const pixelId = 'PIXEL-456';

    const { result } = renderHook(() => useAnalytics(), {
      wrapper: ({ children }) =>
        React.createElement(AnalyticsProvider, {
          measurementId,
          pixelId,
          children,
        }),
    });

    expect(result.current).toBeDefined();
    expect(result.current.trackBusinessCreated).toBeDefined();
    expect(result.current.trackSubscriptionCreated).toBeDefined();
    expect(result.current.trackPaymentReceived).toBeDefined();
  });

  it('should provide all tracking methods', () => {
    const { result } = renderHook(() => useAnalytics(), {
      wrapper: ({ children }) =>
        React.createElement(AnalyticsProvider, {
          measurementId: 'G-TEST123',
          pixelId: 'PIXEL-456',
          children,
        }),
    });

    expect(typeof result.current.trackBusinessCreated).toBe('function');
    expect(typeof result.current.trackSubscriptionCreated).toBe('function');
    expect(typeof result.current.trackPaymentReceived).toBe('function');
    expect(typeof result.current.trackValidationSuccess).toBe('function');
    expect(typeof result.current.trackMetaPixelPurchase).toBe('function');
    expect(typeof result.current.trackMetaPixelLead).toBe('function');
    expect(typeof result.current.trackMetaPixelViewContent).toBe('function');
  });
});
