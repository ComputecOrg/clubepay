import { useAnalyticsContext } from '@/lib/analyticsContext';

/**
 * Hook to access Analytics instance for tracking events
 *
 * Usage:
 * const analytics = useAnalytics();
 * analytics.trackBusinessCreated(userId, businessName, industry);
 */
export function useAnalytics() {
  return useAnalyticsContext();
}
