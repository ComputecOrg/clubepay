import React, { createContext, useContext, ReactNode } from 'react';
import { Analytics } from './analytics';

const AnalyticsContext = createContext<Analytics | null>(null);

interface AnalyticsProviderProps {
  children: ReactNode;
  measurementId?: string;
  pixelId?: string;
}

/**
 * Provider component that initializes and provides Analytics instance to child components
 */
export function AnalyticsProvider({
  children,
  measurementId = process.env.NEXT_PUBLIC_GA_ID || 'G-PLACEHOLDER',
  pixelId = process.env.NEXT_PUBLIC_PIXEL_ID || 'PIXEL-PLACEHOLDER',
}: AnalyticsProviderProps) {
  const analyticsInstance = React.useMemo(() => {
    // Only initialize if we have valid IDs (not placeholders in prod)
    const isProduction = process.env.NODE_ENV === 'production';
    const hasValidGA = isProduction ? !measurementId.includes('PLACEHOLDER') : true;
    const hasValidPixel = isProduction ? !pixelId.includes('PLACEHOLDER') : true;

    if (!hasValidGA && !hasValidPixel) {
      console.warn('Analytics IDs not configured. Using no-op instance.');
      // Return a no-op instance for development without IDs
      return new Analytics('G-NOOP', 'PIXEL-NOOP');
    }

    try {
      return new Analytics(measurementId, pixelId);
    } catch (e) {
      console.error('Failed to initialize Analytics:', e);
      // Return no-op instance on error
      return new Analytics('G-NOOP', 'PIXEL-NOOP');
    }
  }, [measurementId, pixelId]);

  return (
    <AnalyticsContext.Provider value={analyticsInstance}>
      {children}
    </AnalyticsContext.Provider>
  );
}

/**
 * Hook to access Analytics instance from context
 */
export function useAnalyticsContext(): Analytics {
  const analytics = useContext(AnalyticsContext);
  if (!analytics) {
    throw new Error('useAnalytics must be used within AnalyticsProvider');
  }
  return analytics;
}
