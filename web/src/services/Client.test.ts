import axios from "axios";
import { vi } from "vitest";

import { hasServiceError, toData, toDataRateLimited } from "@services/Api";
import * as Client from "@services/Client";

vi.mock("axios");
vi.mock("@services/Api");

it("handles successful post", async () => {
    const mockRes = { status: 200, data: { status: "OK", data: "test" } };
    (axios.post as any).mockResolvedValue(mockRes);
    (hasServiceError as any).mockReturnValue({ errored: false });
    (toData as any).mockReturnValue("test");

    const result = await Client.Post("/path", {});
    expect(result).toBe("test");
});

it("handles successful post with optional response", async () => {
    const mockRes = { status: 200, data: { status: "OK", data: "test" } };
    (axios.post as any).mockResolvedValue(mockRes);
    (hasServiceError as any).mockReturnValue({ errored: false });
    (toData as any).mockReturnValue("test");

    const result = await Client.PostWithOptionalResponse("/path", {});
    expect(result).toBe("test");
});

it("handles successful rate limited post", async () => {
    const mockRes = { status: 200, data: { status: "OK", data: "test" } };
    (axios.post as any).mockResolvedValue(mockRes);
    (hasServiceError as any).mockReturnValue({ errored: false });
    (toDataRateLimited as any).mockReturnValue({ limited: false, retryAfter: 0, data: "test" });

    const result = await Client.PostWithOptionalResponseRateLimited("/path", {});
    expect(result).toEqual({ limited: false, retryAfter: 0, data: "test" });
});

it("handles rate limited post", async () => {
    const mockRes = { status: 429, data: { status: "KO" }, headers: { "retry-after": "30" } };
    (axios.post as any).mockResolvedValue(mockRes);
    (hasServiceError as any).mockReturnValue({ errored: true });
    (toDataRateLimited as any).mockReturnValue({ limited: true, retryAfter: 30 });

    const result = await Client.PostWithOptionalResponseRateLimited("/path", {});
    expect(result).toEqual({ limited: true, retryAfter: 30 });
});

it("throws on post error", async () => {
    const mockRes = { status: 400, data: { status: "KO", message: "error" } };
    (axios.post as any).mockResolvedValue(mockRes);
    (hasServiceError as any).mockReturnValue({ errored: true, message: "error" });

    await expect(Client.PostWithOptionalResponse("/path", {})).rejects.toThrow(
        "Failed POST to /path. Code: 400. Message: error",
    );
});

it("throws on rate limited post error", async () => {
    const mockRes = { status: 400, data: { status: "KO", message: "error" } };
    (axios.post as any).mockResolvedValue(mockRes);
    (hasServiceError as any).mockReturnValue({ errored: true, message: "error" });

    await expect(Client.PostWithOptionalResponseRateLimited("/path", {})).rejects.toThrow(
        "Failed POST to /path. Code: 400. Message: error",
    );
});

it("throws on post with no data", async () => {
    const mockRes = { status: 200, data: { status: "OK" } };
    (axios.post as any).mockResolvedValue(mockRes);
    (hasServiceError as any).mockReturnValue({ errored: false });
    (toData as any).mockReturnValue(undefined);

    await expect(Client.Post("/path", {})).rejects.toThrow("unexpected type of response");
});

it("handles successful get", async () => {
    const mockRes = { status: 200, data: { status: "OK", data: "test" } };
    (axios.get as any).mockResolvedValue(mockRes);
    (hasServiceError as any).mockReturnValue({ errored: false });
    (toData as any).mockReturnValue("test");

    const result = await Client.Get("/path");
    expect(result).toBe("test");
});

it("handles get with optional data returning data", async () => {
    const mockRes = { status: 200, data: { status: "OK", data: "test" } };
    (axios.get as any).mockResolvedValue(mockRes);
    (hasServiceError as any).mockReturnValue({ errored: false });
    (toData as any).mockReturnValue("test");

    const result = await Client.GetWithOptionalData("/path");
    expect(result).toBe("test");
});

it("handles get with optional data returning null", async () => {
    const mockRes = { status: 200, data: { status: "OK" } };
    (axios.get as any).mockResolvedValue(mockRes);
    (hasServiceError as any).mockReturnValue({ errored: false });
    (toData as any).mockReturnValue(null);

    const result = await Client.GetWithOptionalData("/path");
    expect(result).toBeNull();
});

it("throws on get with no data", async () => {
    const mockRes = { status: 200, data: { status: "OK" } };
    (axios.get as any).mockResolvedValue(mockRes);
    (hasServiceError as any).mockReturnValue({ errored: false });
    (toData as any).mockReturnValue(undefined);

    await expect(Client.Get("/path")).rejects.toThrow("unexpected type of response");
});

it("handles successful put with optional response", async () => {
    const mockRes = { status: 200, data: { status: "OK", data: "test" } };
    (axios.put as any).mockResolvedValue(mockRes);
    (hasServiceError as any).mockReturnValue({ errored: false });
    (toData as any).mockReturnValue("test");

    const result = await Client.PutWithOptionalResponse("/path", {});
    expect(result).toBe("test");
});

it("throws on put error", async () => {
    const mockRes = { status: 400, data: { status: "KO", message: "error" } };
    (axios.put as any).mockResolvedValue(mockRes);
    (hasServiceError as any).mockReturnValue({ errored: true, message: "error" });

    await expect(Client.PutWithOptionalResponse("/path", {})).rejects.toThrow(
        "Failed POST to /path. Code: 400. Message: error",
    );
});

it("handles successful delete with optional response", async () => {
    const mockRes = { status: 200, data: { status: "OK", data: "test" } };
    (axios.delete as any).mockResolvedValue(mockRes);
    (hasServiceError as any).mockReturnValue({ errored: false });
    (toData as any).mockReturnValue("test");

    const result = await Client.DeleteWithOptionalResponse("/path", {});
    expect(result).toBe("test");
});

it("throws on delete error", async () => {
    const mockRes = { status: 400, data: { status: "KO", message: "error" } };
    (axios.delete as any).mockResolvedValue(mockRes);
    (hasServiceError as any).mockReturnValue({ errored: true, message: "error" });

    await expect(Client.DeleteWithOptionalResponse("/path", {})).rejects.toThrow(
        "Failed DELETE to /path. Code: 400. Message: error",
    );
});

it("handles successful put", async () => {
    const mockRes = { status: 200, data: { status: "OK", data: "test" } };
    (axios.put as any).mockResolvedValue(mockRes);
    (hasServiceError as any).mockReturnValue({ errored: false });
    (toData as any).mockReturnValue("test");

    const result = await Client.Put("/path", {});
    expect(result).toBe("test");
});

it("throws on put with no data", async () => {
    const mockRes = { status: 200, data: { status: "OK" } };
    (axios.put as any).mockResolvedValue(mockRes);
    (hasServiceError as any).mockReturnValue({ errored: false });
    (toData as any).mockReturnValue(undefined);

    await expect(Client.Put("/path", {})).rejects.toThrow("unexpected type of response");
});

it("throws on get error", async () => {
    const mockRes = { status: 400, data: { status: "KO", message: "error" } };
    (axios.get as any).mockResolvedValue(mockRes);
    (hasServiceError as any).mockReturnValue({ errored: true, message: "error" });

    await expect(Client.Get("/path")).rejects.toThrow("Failed GET from /path. Code: 400.");
});

it("throws on get with optional data error", async () => {
    const mockRes = { status: 400, data: { status: "KO", message: "error" } };
    (axios.get as any).mockResolvedValue(mockRes);
    (hasServiceError as any).mockReturnValue({ errored: true, message: "error" });

    await expect(Client.GetWithOptionalData("/path")).rejects.toThrow("Failed GET from /path. Code: 400.");
});
