import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";

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

// Mock QRCodeSVG since it relies on canvas/SVG internals
vi.mock("qrcode.react", () => ({
  QRCodeSVG: ({ value }: { value: string }) => (
    <div data-testid="qrcode">{value}</div>
  ),
}));

beforeEach(() => {
  vi.clearAllMocks();
});

import DashboardPage from "@/app/(auth)/dashboard/page";

const businessResponse = {
  business: {
    id: "1",
    name: "Cafe Central",
    slug: "cafe-central",
    segment: "cafeteria",
    subscriber_count: 3,
  },
};

const subscriptionsResponse = {
  subscriptions: [
    {
      id: "s1",
      subscriber_name: "João",
      subscriber_email: "joao@test.com",
      plan_name: "Diário",
      status: "active",
    },
  ],
};

function mockApiSuccess() {
  mockGetToken.mockReturnValue("fake-token");
  mockApi.get.mockImplementation((path: string) => {
    if (path === "/api/business") return Promise.resolve(businessResponse);
    if (path === "/api/subscriptions")
      return Promise.resolve(subscriptionsResponse);
    return Promise.reject(new Error("Unknown path"));
  });
}

describe("DashboardPage", () => {
  it("redirects to login when no token", async () => {
    mockGetToken.mockReturnValue(null);

    render(<DashboardPage />);

    await waitFor(() => {
      expect(mockPush).toHaveBeenCalledWith("/login");
    });
  });

  it("renders dashboard with business data and subscriptions", async () => {
    mockApiSuccess();

    render(<DashboardPage />);

    await waitFor(() => {
      expect(screen.getByText("Cafe Central")).toBeInTheDocument();
    });

    expect(screen.getByText("assinapix.com.br/cafe-central")).toBeInTheDocument();

    // Subscriber appears in the list
    expect(screen.getByText("João")).toBeInTheDocument();
    expect(screen.getByText("Diário")).toBeInTheDocument();
    expect(screen.getByText("Ativo")).toBeInTheDocument();

    // StatsCard values
    expect(screen.getByText("Assinantes ativos")).toBeInTheDocument();
    expect(screen.getByText("1")).toBeInTheDocument();
    expect(screen.getByText("1 ativo")).toBeInTheDocument();
    expect(screen.getByText("100%")).toBeInTheDocument();

    // API was called with the token
    expect(mockApi.get).toHaveBeenCalledWith("/api/business", "fake-token");
    expect(mockApi.get).toHaveBeenCalledWith(
      "/api/subscriptions",
      "fake-token"
    );
  });

  it('shows "Carregando..." while loading', () => {
    mockGetToken.mockReturnValue("fake-token");
    // Never resolve the API calls so loading state persists
    mockApi.get.mockReturnValue(new Promise(() => {}));

    render(<DashboardPage />);

    expect(screen.getByText("Carregando...")).toBeInTheDocument();
  });

  it("redirects on 401 error", async () => {
    mockGetToken.mockReturnValue("fake-token");

    // Import the mocked ApiError dynamically so it matches the mock
    const { ApiError } = await import("@/lib/api");
    mockApi.get.mockRejectedValue(new ApiError(401, "Unauthorized"));

    render(<DashboardPage />);

    await waitFor(() => {
      expect(mockClearToken).toHaveBeenCalled();
      expect(mockPush).toHaveBeenCalledWith("/login");
    });
  });

  it("shows error message on generic error", async () => {
    mockGetToken.mockReturnValue("fake-token");
    mockApi.get.mockRejectedValue(new Error("Network error"));

    render(<DashboardPage />);

    await waitFor(() => {
      expect(
        screen.getByText("Erro ao carregar dados.")
      ).toBeInTheDocument();
    });
  });

  it("shows upgrade banner when activeCount >= 12", async () => {
    mockGetToken.mockReturnValue("fake-token");

    const manyActiveSubscriptions = Array.from({ length: 13 }, (_, i) => ({
      id: `s${i}`,
      subscriber_name: `Assinante ${i}`,
      subscriber_email: `sub${i}@test.com`,
      plan_name: "Diário",
      status: "active",
    }));

    mockApi.get.mockImplementation((path: string) => {
      if (path === "/api/business") return Promise.resolve(businessResponse);
      if (path === "/api/subscriptions")
        return Promise.resolve({ subscriptions: manyActiveSubscriptions });
      return Promise.reject(new Error("Unknown path"));
    });

    render(<DashboardPage />);

    await waitFor(() => {
      expect(
        screen.getByText(/Faça upgrade para continuar crescendo/)
      ).toBeInTheDocument();
    });
  });

  it('logout button clears token and redirects to "/login"', async () => {
    mockApiSuccess();

    render(<DashboardPage />);

    await waitFor(() => {
      expect(screen.getByText("Cafe Central")).toBeInTheDocument();
    });

    const logoutButton = screen.getByText("Sair");
    fireEvent.click(logoutButton);

    expect(mockClearToken).toHaveBeenCalled();
    expect(mockPush).toHaveBeenCalledWith("/login");
  });
});
