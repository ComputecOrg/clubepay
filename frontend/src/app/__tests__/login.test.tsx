import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";

// Mock next/navigation
const mockPush = vi.fn();
const mockBack = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: mockPush, back: mockBack }),
  useSearchParams: () => new URLSearchParams(),
}));

// Mock API
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

// Mock auth
vi.mock("@/lib/auth", () => ({
  getToken: vi.fn(),
  setToken: vi.fn(),
  setRole: vi.fn(),
  clearToken: vi.fn(),
}));

import LoginPage from "@/app/(auth)/login/page";
import { api, ApiError } from "@/lib/api";
import { setToken } from "@/lib/auth";

describe("LoginPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders login form with email and password fields", () => {
    render(<LoginPage />);

    expect(screen.getByLabelText("E-mail")).toBeInTheDocument();
    expect(screen.getByLabelText("Senha")).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Entrar" })
    ).toBeInTheDocument();
  });

  it("submits form and redirects to dashboard on success", async () => {
    // Backend retorna user sem token (JWT vem via cookie HttpOnly)
    vi.mocked(api.post).mockResolvedValueOnce({ user: { role: "owner" } });

    render(<LoginPage />);

    fireEvent.change(screen.getByLabelText("E-mail"), {
      target: { value: "user@test.com" },
    });
    fireEvent.change(screen.getByLabelText("Senha"), {
      target: { value: "password123" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Entrar" }));

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith("/api/auth/login", {
        email: "user@test.com",
        password: "password123",
      });
    });

    // setToken sem argumento — JWT gerenciado pelo backend via HttpOnly cookie
    expect(setToken).toHaveBeenCalledWith();
    expect(mockPush).toHaveBeenCalledWith("/dashboard");
  });

  it("shows error message on ApiError", async () => {
    vi.mocked(api.post).mockRejectedValueOnce(
      new ApiError(401, "Credenciais inválidas")
    );

    render(<LoginPage />);

    fireEvent.change(screen.getByLabelText("E-mail"), {
      target: { value: "user@test.com" },
    });
    fireEvent.change(screen.getByLabelText("Senha"), {
      target: { value: "wrongpass" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Entrar" }));

    await waitFor(() => {
      expect(screen.getByText("Credenciais inválidas")).toBeInTheDocument();
    });

    expect(mockPush).not.toHaveBeenCalled();
  });

  it("shows generic error on unknown error", async () => {
    vi.mocked(api.post).mockRejectedValueOnce(new Error("Network failure"));

    render(<LoginPage />);

    fireEvent.change(screen.getByLabelText("E-mail"), {
      target: { value: "user@test.com" },
    });
    fireEvent.change(screen.getByLabelText("Senha"), {
      target: { value: "password123" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Entrar" }));

    await waitFor(() => {
      expect(
        screen.getByText("Erro ao fazer login. Tente novamente.")
      ).toBeInTheDocument();
    });

    expect(mockPush).not.toHaveBeenCalled();
  });

  it("disables submit button while loading", async () => {
    let resolvePost: (value: unknown) => void;
    vi.mocked(api.post).mockImplementationOnce(
      () => new Promise((resolve) => { resolvePost = resolve; })
    );

    render(<LoginPage />);

    fireEvent.change(screen.getByLabelText("E-mail"), {
      target: { value: "user@test.com" },
    });
    fireEvent.change(screen.getByLabelText("Senha"), {
      target: { value: "password123" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Entrar" }));

    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Entrando..." })).toBeDisabled();
    });

    resolvePost!({ token: "jwt-token-123" });

    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Entrar" })).not.toBeDisabled();
    });
  });
});
