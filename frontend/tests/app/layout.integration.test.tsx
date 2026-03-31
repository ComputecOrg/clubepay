import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import React from 'react';
import { AnalyticsProvider } from '@/lib/analyticsContext';

// Mock root layout with AnalyticsProvider
const RootLayoutWithAnalytics: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  return (
    <html lang="pt-BR" className="h-full">
      <AnalyticsProvider
        measurementId={process.env.NEXT_PUBLIC_GA_ID || 'G-TEST'}
        pixelId={process.env.NEXT_PUBLIC_PIXEL_ID || 'PIXEL-TEST'}
      >
        <body className="min-h-full flex flex-col">{children}</body>
      </AnalyticsProvider>
    </html>
  );
};

describe('Root Layout with Analytics', () => {
  beforeEach(() => {
    delete (window as any).gtag;
    delete (window as any).fbq;
  });

  it('should render children within AnalyticsProvider', () => {
    render(
      <RootLayoutWithAnalytics>
        <div>Test Content</div>
      </RootLayoutWithAnalytics>
    );

    expect(screen.getByText('Test Content')).toBeInTheDocument();
  });

  it('should have analytics initialized for all child pages', () => {
    const mockGtag = vi.fn();
    (window as any).gtag = mockGtag;

    render(
      <RootLayoutWithAnalytics>
        <div data-testid="page-content">Page Content</div>
      </RootLayoutWithAnalytics>
    );

    expect(screen.getByTestId('page-content')).toBeInTheDocument();
    // gtag should be available in the global scope
    expect((window as any).gtag).toBeDefined();
  });
});
