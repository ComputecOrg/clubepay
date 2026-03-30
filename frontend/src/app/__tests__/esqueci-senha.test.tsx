import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";

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

beforeEach(() => {
  vi.clearAllMocks();
});

import EsqueciSenhaPage from "@/app/(auth)/esqueci-senha/page";
import { api, ApiError } from "@/lib/api";

describe("EsqueciSenhaPage", () => {
  it("renders form with email field", () => {
    render(<EsqueciSenhaPage />);

    expect(screen.getByLabelText("E-mail")).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Enviar link de recuperação" })
    ).toBeInTheDocument();
  });

  it("shows success message after submit", async () => {
    vi.mocked(api.post).mockResolvedValueOnce({});

    render(<EsqueciSenhaPage />);

    fireEvent.change(screen.getByLabelText("E-mail"), {
      target: { value: "user@test.com" },
    });
    fireEvent.click(
      screen.getByRole("button", { name: "Enviar link de recuperação" })
    );

    await waitFor(() => {
      expect(
        screen.getByText("Email enviado! Verifique sua caixa de entrada.")
      ).toBeInTheDocument();
    });

    expect(api.post).toHaveBeenCalledWith(
      "/api/auth/request-password-reset",
      { email: "user@test.com" }
    );
  });

  it("shows error message on ApiError", async () => {
    vi.mocked(api.post).mockRejectedValueOnce(
      new ApiError(422, "E-mail não encontrado")
    );

    render(<EsqueciSenhaPage />);

    fireEvent.change(screen.getByLabelText("E-mail"), {
      target: { value: "unknown@test.com" },
    });
    fireEvent.click(
      screen.getByRole("button", { name: "Enviar link de recuperação" })
    );

    await waitFor(() => {
      expect(
        screen.getByText("E-mail não encontrado")
      ).toBeInTheDocument();
    });
  });

  it("shows generic error on unknown error", async () => {
    vi.mocked(api.post).mockRejectedValueOnce(new Error("Network failure"));

    render(<EsqueciSenhaPage />);

    fireEvent.change(screen.getByLabelText("E-mail"), {
      target: { value: "user@test.com" },
    });
    fireEvent.click(
      screen.getByRole("button", { name: "Enviar link de recuperação" })
    );

    await waitFor(() => {
      expect(
        screen.getByText("Erro ao enviar e-mail. Tente novamente.")
      ).toBeInTheDocument();
    });
  });
});
