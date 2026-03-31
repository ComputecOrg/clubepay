import { describe, it, expect, vi } from "vitest";

// Mock the analytics module
vi.mock("@/lib/analytics", () => ({
  initMetaPixel: vi.fn(),
}));

describe("MetaPixel Component", () => {
  it("should be defined and importable", async () => {
    // The component should exist and be importable without errors
    // Actual Meta Pixel functionality is tested in src/lib/__tests__/analytics.test.ts
    const MetaPixel = await import("@/components/MetaPixel");
    expect(MetaPixel).toBeDefined();
    expect(MetaPixel.default).toBeDefined();
  });
});
