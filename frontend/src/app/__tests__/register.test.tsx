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
  clearToken: vi.fn(),
}));

import RegisterPage from "@/app/(auth)/register/page";
import { api, ApiError } from "@/lib/api";
import { setToken } from "@/lib/auth";

function fillRegisterForm() {
  fireEvent.change(screen.getByLabelText("Seu nome"), {
    target: { value: "João Silva" },
  });
  fireEvent.change(screen.getByLabelText("E-mail"), {
    target: { value: "joao@test.com" },
  });
  fireEvent.change(screen.getByLabelText("Senha"), {
    target: { value: "senha12345" },
  });
  fireEvent.change(screen.getByLabelText("Nome do negócio"), {
    target: { value: "Café Central" },
  });
  fireEvent.change(screen.getByLabelText("Segmento"), {
    target: { value: "cafeteria" },
  });
}

describe("RegisterPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders registration form with all fields", () => {
    render(<RegisterPage />);

    expect(screen.getByLabelText("Seu nome")).toBeInTheDocument();
    expect(screen.getByLabelText("E-mail")).toBeInTheDocument();
    expect(screen.getByLabelText("Senha")).toBeInTheDocument();
    expect(screen.getByLabelText("Nome do negócio")).toBeInTheDocument();
    expect(screen.getByLabelText("Segmento")).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Criar meu clube" })
    ).toBeInTheDocument();
  });

  it("submits form and redirects to dashboard on success", async () => {
    vi.mocked(api.post).mockResolvedValueOnce({ token: "jwt-token-456" });

    render(<RegisterPage />);
    fillRegisterForm();
    fireEvent.click(screen.getByRole("button", { name: "Criar meu clube" }));

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith("/api/auth/register", {
        name: "João Silva",
        email: "joao@test.com",
        password: "senha12345",
        business_name: "Café Central",
        segment: "cafeteria",
      });
    });

    expect(setToken).toHaveBeenCalledWith("jwt-token-456");
    expect(mockPush).toHaveBeenCalledWith("/dashboard");
  });

  it("shows error message on ApiError", async () => {
    vi.mocked(api.post).mockRejectedValueOnce(
      new ApiError(409, "E-mail já cadastrado")
    );

    render(<RegisterPage />);
    fillRegisterForm();
    fireEvent.click(screen.getByRole("button", { name: "Criar meu clube" }));

    await waitFor(() => {
      expect(screen.getByText("E-mail já cadastrado")).toBeInTheDocument();
    });

    expect(mockPush).not.toHaveBeenCalled();
  });

  it("shows generic error on unknown error", async () => {
    vi.mocked(api.post).mockRejectedValueOnce(new Error("Network failure"));

    render(<RegisterPage />);
    fillRegisterForm();
    fireEvent.click(screen.getByRole("button", { name: "Criar meu clube" }));

    await waitFor(() => {
      expect(
        screen.getByText("Erro ao criar conta. Tente novamente.")
      ).toBeInTheDocument();
    });

    expect(mockPush).not.toHaveBeenCalled();
  });
});
