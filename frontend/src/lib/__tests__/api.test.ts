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

  it("sends authorization header when token provided", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({}),
    });

    await api.get("/api/test", "my-token");
    const callArgs = mockFetch.mock.calls[0];
    expect(callArgs[1].headers["Authorization"]).toBe("Bearer my-token");
  });

  it("does not send authorization header when no token", async () => {
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
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 500,
      statusText: "Internal Server Error",
      json: () => Promise.reject(new Error("invalid json")),
    });

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
});

describe("api.put", () => {
  it("makes PUT request with body and token", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ updated: true }),
    });

    await api.put("/api/items/1", { name: "updated" }, "token");
    const callArgs = mockFetch.mock.calls[0];
    expect(callArgs[1].method).toBe("PUT");
    expect(callArgs[1].body).toBe(JSON.stringify({ name: "updated" }));
    expect(callArgs[1].headers["Authorization"]).toBe("Bearer token");
  });
});

describe("api.del", () => {
  it("makes DELETE request with token", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({}),
    });

    await api.del("/api/items/1", "token");
    const callArgs = mockFetch.mock.calls[0];
    expect(callArgs[1].method).toBe("DELETE");
    expect(callArgs[1].headers["Authorization"]).toBe("Bearer token");
  });

  it("does not send body on DELETE", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({}),
    });

    await api.del("/api/items/1", "token");
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
