/**
 * Meta Pixel (Facebook Pixel) integration for ClubePay
 * Tracks user events for Meta Ads conversion optimization
 */

export enum MetaPixelEvents {
  ViewContent = "ViewContent",
  Lead = "Lead",
  AddToCart = "AddToCart",
  Purchase = "Purchase",
  InitiateCheckout = "InitiateCheckout",
  CompleteRegistration = "CompleteRegistration",
  Subscribe = "Subscribe",
}

interface MetaPixelEventData {
  value?: number;
  currency?: string;
  content_name?: string;
  content_id?: string;
  content_type?: string;
  [key: string]: any;
}

declare global {
  interface Window {
    fbq?: ((type: string, event: string, data?: any) => void) & {
      queue?: any[];
      loaded?: boolean;
      version?: string;
    };
    facebook_ids?: any[];
  }
}

/**
 * Initialize Meta Pixel with the given pixel ID
 * Must be called once on app initialization
 */
export function initMetaPixel(pixelId: string): void {
  // Return early if fbq is already initialized
  if (window.fbq?.loaded) {
    return;
  }

  // Create fbq function if it doesn't exist
  if (!window.fbq) {
    const fbq = function () {
      fbq.queue = fbq.queue || [];
      fbq.queue.push(arguments);
    } as any;
    fbq.push = fbq;
    fbq.loaded = false;
    fbq.version = "2.0";
    fbq.queue = [];
    window.fbq = fbq;

    // Load the Meta Pixel script
    const script = document.createElement("script");
    script.async = true;
    script.src = "https://connect.facebook.net/en_US/fbevents.js";
    const firstScript = document.getElementsByTagName("script")[0];
    if (firstScript && firstScript.parentNode) {
      firstScript.parentNode.insertBefore(script, firstScript);
    }
  }

  // Initialize fbq with pixel ID
  if (window.fbq) {
    window.fbq("init", pixelId);
    window.fbq("track", "PageView");
    window.fbq.loaded = true;
  }
}

/**
 * Track a custom event in Meta Pixel
 * @param eventName - Name of the event to track
 * @param data - Optional event data (value, currency, etc.)
 */
export function trackEvent(
  eventName: string | MetaPixelEvents,
  data?: MetaPixelEventData
): void {
  if (!window.fbq) {
    // Meta Pixel not initialized, silently return
    return;
  }

  window.fbq("track", eventName, data);
}

/**
 * Convenience functions for common events
 */
export const trackViewContent = (contentName: string, contentId?: string) => {
  trackEvent(MetaPixelEvents.ViewContent, { content_name: contentName, content_id: contentId });
};

export const trackLead = (value?: number, currency?: string) => {
  trackEvent(MetaPixelEvents.Lead, { value, currency });
};

export const trackPurchase = (
  value: number,
  currency: string,
  contentName?: string
) => {
  trackEvent(MetaPixelEvents.Purchase, {
    value,
    currency,
    content_name: contentName,
  });
};

export const trackSubscribe = (value?: number, currency?: string) => {
  trackEvent(MetaPixelEvents.Subscribe, { value, currency });
};
