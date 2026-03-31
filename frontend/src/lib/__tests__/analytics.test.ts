import { describe, it, expect, beforeEach, vi } from "vitest";
import { initMetaPixel, trackEvent, MetaPixelEvents } from "@/lib/analytics";

describe("analytics", () => {
  describe("initMetaPixel", () => {
    let fbqMock: any;

    beforeEach(() => {
      // Clear window.fbq before each test
      delete (window as any).fbq;
      delete (window as any).facebook_ids;

      // Mock fbq global
      fbqMock = vi.fn();
      fbqMock.queue = [];
      fbqMock.loaded = false;
      fbqMock.version = "2.0";
      (window as any).fbq = fbqMock;
    });

    it("should load Meta Pixel script if not already loaded", () => {
      delete (window as any).fbq;
      const pixelId = "test-pixel-id-123";

      initMetaPixel(pixelId);

      // Check that fbq is now available
      expect((window as any).fbq).toBeDefined();
    });

    it("should initialize fbq with correct pixel ID", () => {
      const pixelId = "test-pixel-id-456";
      fbqMock.mockClear();

      initMetaPixel(pixelId);

      // First call to fbq should be init with the pixel ID
      expect(fbqMock).toHaveBeenCalledWith("init", pixelId);
    });

    it("should set fbq.loaded after initialization", () => {
      const pixelId = "test-pixel-id-789";
      delete (window as any).fbq;

      initMetaPixel(pixelId);

      expect((window as any).fbq.loaded).toBe(true);
    });

    it("should not reinitialize if fbq is already loaded", () => {
      fbqMock.mockClear();
      fbqMock.loaded = true;

      initMetaPixel("pixel-id");

      // Should not call fbq if already loaded
      expect(fbqMock).not.toHaveBeenCalled();
    });

    it("should track PageView after initialization", () => {
      fbqMock.mockClear();
      const pixelId = "test-pixel-id";

      initMetaPixel(pixelId);

      // Check PageView was tracked
      expect(fbqMock).toHaveBeenCalledWith("track", "PageView");
    });
  });

  describe("trackEvent", () => {
    let fbqMock: any;

    beforeEach(() => {
      delete (window as any).fbq;
      delete (window as any).facebook_ids;

      fbqMock = vi.fn();
      fbqMock.queue = [];
      fbqMock.loaded = true;
      fbqMock.version = "2.0";
      (window as any).fbq = fbqMock;
      fbqMock.mockClear();
    });

    it("should call fbq track with event name", () => {
      trackEvent("ViewContent");

      expect(fbqMock).toHaveBeenCalledWith("track", "ViewContent", undefined);
    });

    it("should call fbq track with event data if provided", () => {
      const eventData = {
        value: 100,
        currency: "BRL",
        content_name: "Test Plan",
      };

      trackEvent("Purchase", eventData);

      expect(fbqMock).toHaveBeenCalledWith("track", "Purchase", eventData);
    });

    it("should handle ViewContent event", () => {
      trackEvent(MetaPixelEvents.ViewContent, {
        content_name: "Test Plan",
      });

      expect(fbqMock).toHaveBeenCalledWith("track", "ViewContent", {
        content_name: "Test Plan",
      });
    });

    it("should handle Lead event", () => {
      trackEvent(MetaPixelEvents.Lead, {
        value: 50,
        currency: "BRL",
      });

      expect(fbqMock).toHaveBeenCalledWith("track", "Lead", {
        value: 50,
        currency: "BRL",
      });
    });

    it("should handle AddToCart event", () => {
      trackEvent(MetaPixelEvents.AddToCart, {
        content_name: "Plan Monthly",
        value: 29.9,
        currency: "BRL",
      });

      expect(fbqMock).toHaveBeenCalledWith("track", "AddToCart", {
        content_name: "Plan Monthly",
        value: 29.9,
        currency: "BRL",
      });
    });

    it("should handle Purchase event", () => {
      trackEvent(MetaPixelEvents.Purchase, {
        value: 100,
        currency: "BRL",
        content_name: "Annual Plan",
      });

      expect(fbqMock).toHaveBeenCalledWith("track", "Purchase", {
        value: 100,
        currency: "BRL",
        content_name: "Annual Plan",
      });
    });

    it("should not throw error if fbq is not initialized", () => {
      delete (window as any).fbq;

      expect(() => trackEvent("ViewContent")).not.toThrow();
    });

    it("should gracefully handle missing fbq", () => {
      (window as any).fbq = undefined;

      const result = trackEvent("Lead");

      // Should return undefined without throwing
      expect(result).toBeUndefined();
    });
  });

  describe("MetaPixelEvents", () => {
    it("should have ViewContent event", () => {
      expect(MetaPixelEvents.ViewContent).toBe("ViewContent");
    });

    it("should have Lead event", () => {
      expect(MetaPixelEvents.Lead).toBe("Lead");
    });

    it("should have AddToCart event", () => {
      expect(MetaPixelEvents.AddToCart).toBe("AddToCart");
    });

    it("should have Purchase event", () => {
      expect(MetaPixelEvents.Purchase).toBe("Purchase");
    });

    it("should have InitiateCheckout event", () => {
      expect(MetaPixelEvents.InitiateCheckout).toBe("InitiateCheckout");
    });

    it("should have CompleteRegistration event", () => {
      expect(MetaPixelEvents.CompleteRegistration).toBe(
        "CompleteRegistration"
      );
    });

    it("should have Subscribe event", () => {
      expect(MetaPixelEvents.Subscribe).toBe("Subscribe");
    });
  });
});
