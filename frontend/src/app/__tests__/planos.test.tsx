import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";

const { mockPush, mockApi, mockGetToken, mockClearToken } = vi.hoisted(() => ({
  mockPush: vi.fn(),
  mockApi: { get: vi.fn(), post: vi.fn(), put: vi.fn(), del: vi.fn() },
  mockGetToken: vi.fn(),
  mockClearToken: vi.fn(),
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: mockPush, back: vi.fn() }),
}));

vi.mock("@/lib/api", () => ({
  api: mockApi,
  ApiError: class extends Error {
    status: number;
    constructor(s: number, m: string) {
      super(m);
      this.status = s;
      this.name = "ApiError";
    }
  },
}));

vi.mock("@/lib/auth", () => ({
  getToken: (...args: unknown[]) => mockGetToken(...args),
  setToken: vi.fn(),
  clearToken: (...args: unknown[]) => mockClearToken(...args),
}));

beforeEach(() => {
  vi.clearAllMocks();
});

import PlanosPage from "@/app/(auth)/planos/page";

const plansResponse = {
  plans: [
    {
      id: 1,
      name: "Cafe Diario",
      description: "1 cafe por dia",
      price_cents: 2990,
      limit_type: "daily" as const,
      limit_count: 1,
      active: true,
    },
  ],
};

describe("PlanosPage", () => {
  it("redirects to login when no token", async () => {
    mockGetToken.mockReturnValue(null);

    render(<PlanosPage />);

    await waitFor(() => {
      expect(mockPush).toHaveBeenCalledWith("/login");
    });
  });

  it("renders plan list with plan data", async () => {
    mockGetToken.mockReturnValue("fake-token");
    mockApi.get.mockResolvedValue(plansResponse);

    render(<PlanosPage />);

    await waitFor(() => {
      expect(screen.getByText("Cafe Diario")).toBeInTheDocument();
    });

    // Plan description
    expect(screen.getByText("1 cafe por dia")).toBeInTheDocument();

    // Price and limit info rendered together in one element
    // formatPrice(2990) = "R$ 29,90" and limit is "1x por dia"
    expect(screen.getByText(/R\$\s*29,90\/mes/)).toBeInTheDocument();
    expect(screen.getByText(/1x por dia/)).toBeInTheDocument();

    // Action buttons
    expect(screen.getByText("Editar")).toBeInTheDocument();
    expect(screen.getByText("Desativar")).toBeInTheDocument();

    // Page header
    expect(screen.getByText("Meus Planos")).toBeInTheDocument();

    // "Criar novo plano" button visible when form is not shown
    expect(screen.getByText("+ Criar novo plano")).toBeInTheDocument();

    // API was called (auth via HttpOnly cookie, no token param)
    expect(mockApi.get).toHaveBeenCalledWith("/api/plans");
  });

  it("shows empty state when no plans", async () => {
    mockGetToken.mockReturnValue("fake-token");
    mockApi.get.mockResolvedValue({ plans: [] });

    render(<PlanosPage />);

    await waitFor(() => {
      expect(
        screen.getByText("Nenhum plano criado ainda. Crie seu primeiro plano!")
      ).toBeInTheDocument();
    });
  });

  it("shows error message on API error", async () => {
    mockGetToken.mockReturnValue("fake-token");
    mockApi.get.mockRejectedValue(new Error("Network error"));

    render(<PlanosPage />);

    await waitFor(() => {
      expect(
        screen.getByText("Erro ao carregar planos.")
      ).toBeInTheDocument();
    });
  });

  it("redirects on 401 error", async () => {
    mockGetToken.mockReturnValue("fake-token");

    const { ApiError } = await import("@/lib/api");
    mockApi.get.mockRejectedValue(new ApiError(401, "Unauthorized"));

    render(<PlanosPage />);

    await waitFor(() => {
      expect(mockClearToken).toHaveBeenCalled();
      expect(mockPush).toHaveBeenCalledWith("/login");
    });
  });
});
