import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import StatsCard from "@/components/StatsCard";

describe("StatsCard", () => {
  it("renders label and numeric value", () => {
    render(<StatsCard label="Assinantes ativos" value={42} />);
    expect(screen.getByText("Assinantes ativos")).toBeDefined();
    expect(screen.getByText("42")).toBeDefined();
  });

  it("renders string values", () => {
    render(<StatsCard label="MRR" value="R$ 150,00" />);
    expect(screen.getByText("MRR")).toBeDefined();
    expect(screen.getByText("R$ 150,00")).toBeDefined();
  });

  it("renders zero value", () => {
    render(<StatsCard label="Cancelamentos" value={0} />);
    expect(screen.getByText("Cancelamentos")).toBeDefined();
    expect(screen.getByText("0")).toBeDefined();
  });
});
