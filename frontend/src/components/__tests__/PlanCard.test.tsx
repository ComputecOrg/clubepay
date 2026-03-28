import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import PlanCard from "@/components/PlanCard";

// Mock next/link to render a plain anchor
vi.mock("next/link", () => ({
  default: ({
    href,
    children,
    ...props
  }: {
    href: string;
    children: React.ReactNode;
    [key: string]: unknown;
  }) => (
    <a href={href} {...props}>
      {children}
    </a>
  ),
}));

const basePlan = {
  id: "plan-1",
  name: "Plano Café",
  description: "Um café por dia",
  price_cents: 4990,
  limit_type: "daily" as const,
  limit_value: 1,
};

describe("PlanCard", () => {
  it("renders plan name", () => {
    render(<PlanCard plan={basePlan} slug="meu-cafe" />);
    expect(screen.getByText("Plano Café")).toBeDefined();
  });

  it("renders plan description", () => {
    render(<PlanCard plan={basePlan} slug="meu-cafe" />);
    expect(screen.getByText("Um café por dia")).toBeDefined();
  });

  it("renders formatted price", () => {
    render(<PlanCard plan={basePlan} slug="meu-cafe" />);
    // 4990 cents = R$ 49,90
    expect(screen.getByText(/49,90/)).toBeDefined();
  });

  it("renders daily limit label", () => {
    render(<PlanCard plan={basePlan} slug="meu-cafe" />);
    expect(screen.getByText("1x por dia")).toBeDefined();
  });

  it("renders monthly limit label", () => {
    const monthlyPlan = {
      ...basePlan,
      limit_type: "monthly" as const,
      limit_value: 4,
    };
    render(<PlanCard plan={monthlyPlan} slug="meu-cafe" />);
    expect(screen.getByText("4x por mês")).toBeDefined();
  });

  it("renders subscribe link with correct href", () => {
    render(<PlanCard plan={basePlan} slug="meu-cafe" />);
    const link = screen.getByText("Assinar agora");
    expect(link.closest("a")?.getAttribute("href")).toBe(
      "/assinar/meu-cafe?plan=plan-1"
    );
  });

  it("does not render description when empty", () => {
    const noPlan = { ...basePlan, description: "" };
    render(<PlanCard plan={noPlan} slug="meu-cafe" />);
    // The description paragraph should not exist
    expect(screen.queryByText("Um café por dia")).toBeNull();
  });
});
