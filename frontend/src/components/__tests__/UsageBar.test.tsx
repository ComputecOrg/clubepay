import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import UsageBar from "@/components/UsageBar";

describe("UsageBar", () => {
  it("renders daily usage with correct label", () => {
    render(<UsageBar used={1} limit={3} period="daily" />);
    expect(screen.getByText("1/3 hoje")).toBeDefined();
  });

  it("renders monthly usage with correct label", () => {
    render(<UsageBar used={2} limit={4} period="monthly" />);
    expect(screen.getByText("2/4 este mês")).toBeDefined();
  });

  it("renders percentage", () => {
    render(<UsageBar used={1} limit={2} period="daily" />);
    expect(screen.getByText("50%")).toBeDefined();
  });

  it("caps percentage at 100%", () => {
    render(<UsageBar used={5} limit={3} period="daily" />);
    expect(screen.getByText("100%")).toBeDefined();
  });

  it("renders 0% when limit is 0", () => {
    render(<UsageBar used={0} limit={0} period="daily" />);
    expect(screen.getByText("0%")).toBeDefined();
  });

  it("renders progress bar with correct width", () => {
    const { container } = render(<UsageBar used={3} limit={4} period="monthly" />);
    const bar = container.querySelector("[style]");
    expect(bar).not.toBeNull();
    expect(bar?.getAttribute("style")).toContain("width: 75%");
  });
});
