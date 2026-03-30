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

import PerfilPage from "@/app/(auth)/perfil/page";
import { api, ApiError } from "@/lib/api";
import { getToken, clearToken } from "@/lib/auth";

const profileResponse = {
  user: {
    id: "1",
    email: "test@email.com",
    name: "Test User",
    phone: "11999999999",
    role: "owner",
  },
};

describe("PerfilPage", () => {
  it("redirects to login when no token", async () => {
    vi.mocked(getToken).mockReturnValue(null);

    render(<PerfilPage />);

    await waitFor(() => {
      expect(mockPush).toHaveBeenCalledWith("/login");
    });
  });

  it("renders profile form with user data", async () => {
    vi.mocked(getToken).mockReturnValue("valid-token");
    vi.mocked(api.get).mockResolvedValue(profileResponse);

    render(<PerfilPage />);

    await waitFor(() => {
      expect(screen.getByDisplayValue("test@email.com")).toBeInTheDocument();
    });

    const emailInput = screen.getByDisplayValue("test@email.com");
    expect(emailInput).toBeDisabled();

    expect(screen.getByDisplayValue("Test User")).toBeInTheDocument();
    expect(screen.getByDisplayValue("11999999999")).toBeInTheDocument();

    expect(
      screen.getByRole("button", { name: "Salvar" })
    ).toBeInTheDocument();

    expect(screen.getByPlaceholderText("Senha atual")).toBeInTheDocument();
    expect(
      screen.getByPlaceholderText("Mínimo 8 caracteres")
    ).toBeInTheDocument();
    expect(
      screen.getByPlaceholderText("Repita a nova senha")
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Alterar senha" })
    ).toBeInTheDocument();
  });

  it("shows 'Carregando...' while loading", () => {
    vi.mocked(getToken).mockReturnValue("valid-token");
    vi.mocked(api.get).mockImplementation(() => new Promise(() => {}));

    render(<PerfilPage />);

    expect(screen.getByText("Carregando...")).toBeInTheDocument();
  });

  it("redirects on 401 error", async () => {
    vi.mocked(getToken).mockReturnValue("expired-token");
    vi.mocked(api.get).mockRejectedValue(new ApiError(401, "Unauthorized"));

    render(<PerfilPage />);

    await waitFor(() => {
      expect(clearToken).toHaveBeenCalled();
      expect(mockPush).toHaveBeenCalledWith("/login");
    });
  });
});
