import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";

const mockPush = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: mockPush, back: vi.fn() }),
  useSearchParams: () => new URLSearchParams(),
}));

vi.mock("@/lib/api", () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    del: vi.fn(),
  },
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
  getToken: vi.fn(),
  setToken: vi.fn(),
  clearToken: vi.fn(),
}));

beforeEach(() => {
  vi.clearAllMocks();
});

import MeuPlanoPage from "@/app/meu-plano/page";
import { api, ApiError } from "@/lib/api";
import { getToken, clearToken } from "@/lib/auth";

const planResponse = {
  plan: {
    id: 1,
    name: "Cafe Diario",
    description: "1 cafe por dia",
    price_cents: 2990,
    limit_type: "daily",
    limit_count: 1,
  },
  business: {
    id: 1,
    name: "Cafe Central",
    slug: "cafe-central",
  },
  subscription: {
    id: 1,
    status: "active",
    period_end: "2026-04-15T00:00:00Z",
  },
};

const referralResponse = { code: "ABC123" };

describe("MeuPlanoPage", () => {
  it("redirects to login when no token", async () => {
    vi.mocked(getToken).mockReturnValue(null);

    render(<MeuPlanoPage />);

    await waitFor(() => {
      expect(mockPush).toHaveBeenCalledWith("/login");
    });
  });

  it("renders plan data with usage bar and referral code", async () => {
    vi.mocked(getToken).mockReturnValue("valid-token");
    vi.mocked(api.get).mockImplementation((url: string) => {
      if (url === "/api/my-plan") return Promise.resolve(planResponse);
      if (url === "/api/my-referral-code")
        return Promise.resolve(referralResponse);
      return Promise.reject(new Error("unexpected url"));
    });

    render(<MeuPlanoPage />);

    await waitFor(() => {
      expect(screen.getByText("Cafe Diario")).toBeInTheDocument();
    });

    expect(screen.getByText("Cafe Central")).toBeInTheDocument();
    expect(screen.getByText(/R\$\s*29,90\/mes/)).toBeInTheDocument();
    expect(screen.getByText("Ativo")).toBeInTheDocument();
    expect(screen.getByText("ABC123")).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Cancelar assinatura" })
    ).toBeInTheDocument();
  });

  it("shows 'Carregando...' while loading", () => {
    vi.mocked(getToken).mockReturnValue("valid-token");
    vi.mocked(api.get).mockImplementation(() => new Promise(() => {}));

    render(<MeuPlanoPage />);

    expect(screen.getByText("Carregando...")).toBeInTheDocument();
  });

  it("redirects on 401 error", async () => {
    vi.mocked(getToken).mockReturnValue("expired-token");
    vi.mocked(api.get).mockRejectedValue(new ApiError(401, "Unauthorized"));

    render(<MeuPlanoPage />);

    await waitFor(() => {
      expect(clearToken).toHaveBeenCalled();
      expect(mockPush).toHaveBeenCalledWith("/login");
    });
  });
});
