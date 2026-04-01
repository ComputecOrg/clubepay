import { describe, it, expect, beforeEach, vi } from 'vitest';
import { Analytics } from '@/lib/analytics';

describe('Analytics', () => {
  beforeEach(() => {
    // Clear window.gtag and window.fbq before each test
    delete (window as any).gtag;
    delete (window as any).fbq;
  });

  describe('GA4 Initialization', () => {
    it('should initialize GA4 with measurement ID', () => {
      const measurementId = 'G-TEST123';
      const analytics = new Analytics(measurementId, 'PIXEL-456');

      expect((window as any).gtag).toBeDefined();
      expect(analytics.getMeasurementId()).toBe(measurementId);
    });

    it('should throw error if measurement ID is empty', () => {
      expect(() => new Analytics('', 'PIXEL-456')).toThrow(
        'GA4 measurement ID is required'
      );
    });
  });

  describe('GA4 Event Tracking', () => {
    it('should track business_created event with correct parameters', () => {
      const mockGtag = vi.fn();
      (window as any).gtag = mockGtag;

      const analytics = new Analytics('G-TEST123', 'PIXEL-456');
      analytics.trackBusinessCreated('user-123', 'My Cafe', 'food');

      expect(mockGtag).toHaveBeenCalledWith('event', 'business_created', {
        user_id: 'user-123',
        business_name: 'My Cafe',
        industry: 'food',
      });
    });

    it('should track subscription_created event', () => {
      const mockGtag = vi.fn();
      (window as any).gtag = mockGtag;

      const analytics = new Analytics('G-TEST123', 'PIXEL-456');
      analytics.trackSubscriptionCreated('sub-456', 'plan-789', 'Premium');

      expect(mockGtag).toHaveBeenCalledWith('event', 'subscription_created', {
        user_id: 'sub-456',
        plan_id: 'plan-789',
        plan_name: 'Premium',
      });
    });

    it('should track payment_received event with amount', () => {
      const mockGtag = vi.fn();
      (window as any).gtag = mockGtag;

      const analytics = new Analytics('G-TEST123', 'PIXEL-456');
      analytics.trackPaymentReceived('user-123', 9900, 'Premium');

      expect(mockGtag).toHaveBeenCalledWith('event', 'purchase', {
        user_id: 'user-123',
        value: 99.00,
        currency: 'BRL',
        plan_type: 'Premium',
      });
    });

    it('should track validation_success event', () => {
      const mockGtag = vi.fn();
      (window as any).gtag = mockGtag;

      const analytics = new Analytics('G-TEST123', 'PIXEL-456');
      analytics.trackValidationSuccess('user-123');

      expect(mockGtag).toHaveBeenCalledWith('event', 'validation_success', {
        user_id: 'user-123',
      });
    });
  });

  describe('Meta Pixel', () => {
    it('should initialize Meta Pixel with pixel ID', () => {
      const pixelId = 'PIXEL-456';
      const analytics = new Analytics('G-TEST123', pixelId);

      expect((window as any).fbq).toBeDefined();
      expect(analytics.getPixelId()).toBe(pixelId);
    });

    it('should throw error if pixel ID is empty', () => {
      expect(() => new Analytics('G-TEST123', '')).toThrow(
        'Meta Pixel ID is required'
      );
    });

    it('should track Purchase event for Meta Pixel', () => {
      const mockFbq = vi.fn();
      (window as any).fbq = mockFbq;

      const analytics = new Analytics('G-TEST123', 'PIXEL-456');
      analytics.trackMetaPixelPurchase(9900, 'BRL', 'sub-123');

      expect(mockFbq).toHaveBeenCalledWith('track', 'Purchase', {
        value: 99.00,
        currency: 'BRL',
        content_name: 'Subscription',
        content_id: 'sub-123',
      });
    });

    it('should track Lead event for Meta Pixel', () => {
      const mockFbq = vi.fn();
      (window as any).fbq = mockFbq;

      const analytics = new Analytics('G-TEST123', 'PIXEL-456');
      analytics.trackMetaPixelLead('Lead registration');

      expect(mockFbq).toHaveBeenCalledWith('track', 'Lead', {
        content_name: 'Lead registration',
      });
    });

    it('should track ViewContent event for Meta Pixel', () => {
      const mockFbq = vi.fn();
      (window as any).fbq = mockFbq;

      const analytics = new Analytics('G-TEST123', 'PIXEL-456');
      analytics.trackMetaPixelViewContent('plan-123', 'Premium Plan');

      expect(mockFbq).toHaveBeenCalledWith('track', 'ViewContent', {
        content_id: 'plan-123',
        content_name: 'Premium Plan',
      });
    });
  });

  describe('Event tracking safety', () => {
    it('should not throw if gtag is not available', () => {
      delete (window as any).gtag;
      const analytics = new Analytics('G-TEST123', 'PIXEL-456');

      expect(() => {
        analytics.trackBusinessCreated('user-123', 'Cafe', 'food');
      }).not.toThrow();
    });

    it('should not throw if fbq is not available', () => {
      delete (window as any).fbq;
      const analytics = new Analytics('G-TEST123', 'PIXEL-456');

      expect(() => {
        analytics.trackMetaPixelPurchase(5000, 'BRL', 'sub-123');
      }).not.toThrow();
    });
  });
});
