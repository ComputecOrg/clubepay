import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import React from 'react';
import { AnalyticsProvider } from '@/lib/analyticsContext';
import { useAnalytics } from '@/hooks/useAnalytics';

// Mock component that tracks events
const MockDashboard: React.FC = () => {
  const analytics = useAnalytics();

  React.useEffect(() => {
    // Track page view
    analytics.trackBusinessCreated('user-123', 'Test Cafe', 'food');
  }, [analytics]);

  return (
    <div>
      <h1>Dashboard</h1>
      <button
        onClick={() => {
          analytics.trackSubscriptionCreated(
            'user-123',
            'plan-456',
            'Premium'
          );
        }}
      >
        Create Subscription
      </button>
    </div>
  );
};

describe('Analytics Integration', () => {
  beforeEach(() => {
    delete (window as any).gtag;
    delete (window as any).fbq;
  });

  it('should track events when component renders', () => {
    const mockGtag = vi.fn();
    (window as any).gtag = mockGtag;

    render(
      <AnalyticsProvider measurementId="G-TEST123" pixelId="PIXEL-456">
        <MockDashboard />
      </AnalyticsProvider>
    );

    expect(screen.getByText('Dashboard')).toBeInTheDocument();
    expect(mockGtag).toHaveBeenCalledWith('event', 'business_created', {
      user_id: 'user-123',
      business_name: 'Test Cafe',
      industry: 'food',
    });
  });

  it('should have button that can trigger analytics', () => {
    const mockGtag = vi.fn();
    (window as any).gtag = mockGtag;

    render(
      <AnalyticsProvider measurementId="G-TEST123" pixelId="PIXEL-456">
        <MockDashboard />
      </AnalyticsProvider>
    );

    const button = screen.getByRole('button', { name: /create subscription/i });
    expect(button).toBeInTheDocument();

    // Simulate button click
    button.click();

    // Should have called gtag for the subscription creation
    const subscriptionCalls = mockGtag.mock.calls.filter(
      (call) => call[1] === 'subscription_created'
    );

    expect(subscriptionCalls.length).toBeGreaterThan(0);
  });

  it('should work with AnalyticsProvider wrapper', () => {
    const mockGtag = vi.fn();
    (window as any).gtag = mockGtag;

    render(
      <AnalyticsProvider measurementId="G-TEST123" pixelId="PIXEL-456">
        <MockDashboard />
      </AnalyticsProvider>
    );

    expect(screen.getByText('Dashboard')).toBeInTheDocument();
  });
});
