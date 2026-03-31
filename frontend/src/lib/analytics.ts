/**
 * Frontend analytics integration for GA4 and Meta Pixel
 */

export class Analytics {
  private measurementId: string;
  private pixelId: string;

  constructor(measurementId: string, pixelId: string) {
    if (!measurementId) {
      throw new Error('GA4 measurement ID is required');
    }
    if (!pixelId) {
      throw new Error('Meta Pixel ID is required');
    }

    this.measurementId = measurementId;
    this.pixelId = pixelId;

    this.initializeGA4();
    this.initializeMetaPixel();
  }

  private initializeGA4(): void {
    if (typeof window === 'undefined') return;

    // Initialize gtag function
    (window as any).gtag = (window as any).gtag || function() {
      const dataLayer = (window as any).dataLayer || [];
      dataLayer.push(arguments);
    };

    // Add GA4 script
    const script = document.createElement('script');
    script.async = true;
    script.src = `https://www.googletagmanager.com/gtag/js?id=${this.measurementId}`;
    document.head.appendChild(script);

    // Configure GA4
    (window as any).gtag('js', new Date());
    (window as any).gtag('config', this.measurementId);
  }

  private initializeMetaPixel(): void {
    if (typeof window === 'undefined') return;

    // Initialize fbq function
    (window as any).fbq =
      (window as any).fbq ||
      function() {
        const args = Array.from(arguments);
        if ((window as any).fbq.callQueue) {
          (window as any).fbq.callQueue.push(args);
        }
      };
    (window as any).fbq.callQueue = (window as any).fbq.callQueue || [];
    (window as any).fbq.version = '2.0';

    // Add Meta Pixel script
    const script = document.createElement('script');
    script.async = true;
    script.src = `https://connect.facebook.net/en_US/fbevents.js`;
    document.head.appendChild(script);

    // Initialize pixel
    (window as any).fbq('init', this.pixelId);
    (window as any).fbq('track', 'PageView');
  }

  getMeasurementId(): string {
    return this.measurementId;
  }

  getPixelId(): string {
    return this.pixelId;
  }

  // GA4 Business Events
  trackBusinessCreated(
    userId: string,
    businessName: string,
    industry: string
  ): void {
    if (typeof window === 'undefined' || !(window as any).gtag) return;

    (window as any).gtag('event', 'business_created', {
      user_id: userId,
      business_name: businessName,
      industry: industry,
    });
  }

  trackSubscriptionCreated(
    userId: string,
    planId: string,
    planName: string,
    referralCode?: string
  ): void {
    if (typeof window === 'undefined' || !(window as any).gtag) return;

    const params: Record<string, string> = {
      user_id: userId,
      plan_id: planId,
      plan_name: planName,
    };

    if (referralCode) {
      params.referral_code = referralCode;
    }

    (window as any).gtag('event', 'subscription_created', params);
  }

  trackPaymentReceived(
    userId: string,
    amountInCents: number,
    planType: string
  ): void {
    if (typeof window === 'undefined' || !(window as any).gtag) return;

    (window as any).gtag('event', 'purchase', {
      user_id: userId,
      value: amountInCents / 100,
      currency: 'BRL',
      plan_type: planType,
    });
  }

  trackValidationSuccess(userId: string): void {
    if (typeof window === 'undefined' || !(window as any).gtag) return;

    (window as any).gtag('event', 'validation_success', {
      user_id: userId,
    });
  }

  // Meta Pixel Events
  trackMetaPixelPurchase(
    amountInCents: number,
    currency: string,
    contentId: string
  ): void {
    if (typeof window === 'undefined' || !(window as any).fbq) return;

    (window as any).fbq('track', 'Purchase', {
      value: amountInCents / 100,
      currency: currency,
      content_name: 'Subscription',
      content_id: contentId,
    });
  }

  trackMetaPixelLead(contentName: string): void {
    if (typeof window === 'undefined' || !(window as any).fbq) return;

    (window as any).fbq('track', 'Lead', {
      content_name: contentName,
    });
  }

  trackMetaPixelViewContent(contentId: string, contentName: string): void {
    if (typeof window === 'undefined' || !(window as any).fbq) return;

    (window as any).fbq('track', 'ViewContent', {
      content_id: contentId,
      content_name: contentName,
    });
  }
}
