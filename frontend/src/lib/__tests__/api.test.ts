import { describe, it, expect, vi, beforeEach } from "vitest";
import { api, ApiError } from "@/lib/api";

const mockFetch = vi.fn();
global.fetch = mockFetch;

beforeEach(() => {
  mockFetch.mockReset();
});

describe("api.get", () => {
  it("makes GET request and returns data", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ data: "test" }),
    });

    const result = await api.get("/api/test");
    expect(result).toEqual({ data: "test" });
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining("/api/test"),
      expect.objectContaining({ method: "GET" })
    );
  });

  it("envia credentials include em todas as requisições", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({}),
    });

    await api.get("/api/test");
    const callArgs = mockFetch.mock.calls[0];
    expect(callArgs[1].credentials).toBe("include");
  });

  it("não envia Authorization header (auth via HttpOnly cookie)", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({}),
    });

    await api.get("/api/test");
    const callArgs = mockFetch.mock.calls[0];
    expect(callArgs[1].headers["Authorization"]).toBeUndefined();
  });

  it("throws ApiError on non-ok response", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 404,
      statusText: "Not Found",
      json: () => Promise.resolve({ message: "not found" }),
    });

    await expect(api.get("/api/missing")).rejects.toThrow(ApiError);
  });

  it("throws ApiError with status code", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 403,
      statusText: "Forbidden",
      json: () => Promise.resolve({ message: "access denied" }),
    });

    try {
      await api.get("/api/secret");
    } catch (e) {
      expect(e).toBeInstanceOf(ApiError);
      expect((e as ApiError).status).toBe(403);
      expect((e as ApiError).message).toBe("access denied");
    }
  });

  it("falls back to statusText when json parse fails", async () => {
    // With retry logic, 500 errors are retried up to MAX_RETRIES (2) times = 3 total attempts
    const failResponse = {
      ok: false,
      status: 500,
      statusText: "Internal Server Error",
      json: () => Promise.reject(new Error("invalid json")),
    };
    mockFetch
      .mockResolvedValueOnce(failResponse)
      .mockResolvedValueOnce(failResponse)
      .mockResolvedValueOnce(failResponse);

    try {
      await api.get("/api/broken");
    } catch (e) {
      expect(e).toBeInstanceOf(ApiError);
      expect((e as ApiError).message).toBe("Internal Server Error");
    }
  });
});

describe("api.post", () => {
  it("makes POST request with body", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ id: 1 }),
    });

    const result = await api.post("/api/items", { name: "test" });
    expect(result).toEqual({ id: 1 });

    const callArgs = mockFetch.mock.calls[0];
    expect(callArgs[1].method).toBe("POST");
    expect(callArgs[1].body).toBe(JSON.stringify({ name: "test" }));
  });

  it("sends content-type json header", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({}),
    });

    await api.post("/api/items", { name: "test" });
    const callArgs = mockFetch.mock.calls[0];
    expect(callArgs[1].headers["Content-Type"]).toBe("application/json");
  });

  it("POST envia credentials include", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({}),
    });

    await api.post("/api/items", { name: "test" });
    const callArgs = mockFetch.mock.calls[0];
    expect(callArgs[1].credentials).toBe("include");
  });
});

describe("api.put", () => {
  it("makes PUT request with body via HttpOnly cookie", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ updated: true }),
    });

    await api.put("/api/items/1", { name: "updated" });
    const callArgs = mockFetch.mock.calls[0];
    expect(callArgs[1].method).toBe("PUT");
    expect(callArgs[1].body).toBe(JSON.stringify({ name: "updated" }));
    expect(callArgs[1].credentials).toBe("include");
    expect(callArgs[1].headers["Authorization"]).toBeUndefined();
  });
});

describe("api.del", () => {
  it("makes DELETE request via HttpOnly cookie", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({}),
    });

    await api.del("/api/items/1");
    const callArgs = mockFetch.mock.calls[0];
    expect(callArgs[1].method).toBe("DELETE");
    expect(callArgs[1].credentials).toBe("include");
    expect(callArgs[1].headers["Authorization"]).toBeUndefined();
  });

  it("does not send body on DELETE", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({}),
    });

    await api.del("/api/items/1");
    const callArgs = mockFetch.mock.calls[0];
    expect(callArgs[1].body).toBeUndefined();
  });
});

describe("ApiError", () => {
  it("has correct name and properties", () => {
    const err = new ApiError(422, "validation failed");
    expect(err.name).toBe("ApiError");
    expect(err.status).toBe(422);
    expect(err.message).toBe("validation failed");
    expect(err).toBeInstanceOf(Error);
  });
});

describe("retry logic", () => {
  it("retries on 500 error and succeeds on second attempt", async () => {
    let attempts = 0;
    global.fetch = vi.fn(() => {
      attempts++;
      if (attempts === 1) {
        return Promise.resolve({
          ok: false,
          status: 500,
          statusText: "Internal Server Error",
          json: () => Promise.resolve({ message: "server error" }),
        } as Response);
      }
      return Promise.resolve({
        ok: true,
        json: () => Promise.resolve({ data: "ok" }),
      } as Response);
    });

    const result = await api.get<{ data: string }>("/test");
    expect(result.data).toBe("ok");
    expect(attempts).toBe(2);
  });

  it("does not retry on 4xx errors", async () => {
    let attempts = 0;
    global.fetch = vi.fn(() => {
      attempts++;
      return Promise.resolve({
        ok: false,
        status: 400,
        statusText: "Bad Request",
        json: () => Promise.resolve({ message: "bad request" }),
      } as Response);
    });

    await expect(api.get("/test")).rejects.toThrow();
    expect(attempts).toBe(1);
  });
});
