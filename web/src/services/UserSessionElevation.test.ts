import axios from "axios";
import { vi } from "vitest";

import { hasServiceError, toData, validateStatusOneTimeCode } from "@services/Api";
import {
    deleteUserSessionElevation,
    generateUserSessionElevation,
    getUserSessionElevation,
    verifyUserSessionElevation,
} from "@services/UserSessionElevation";

vi.mock("axios");
vi.mock("@services/Api", () => ({
    UserSessionElevationPath: "/user/elevation",
    hasServiceError: vi.fn(),
    toData: vi.fn(),
    validateStatusOneTimeCode: vi.fn(),
}));

it("gets user session elevation successfully", async () => {
    const mockRes = { status: 200, data: { status: "OK", data: "elevation" } };
    (axios as any).mockResolvedValue(mockRes);
    (hasServiceError as any).mockReturnValue({ errored: false });
    (toData as any).mockReturnValue("elevation");

    const result = await getUserSessionElevation();
    expect(axios).toHaveBeenCalledWith({
        method: "GET",
        url: "/user/elevation",
    });
    expect(result).toBe("elevation");
});

it("gets user session elevation with error", async () => {
    const mockRes = { status: 400, data: { status: "KO", message: "error" } };
    (axios as any).mockResolvedValue(mockRes);
    (hasServiceError as any).mockReturnValue({ errored: true, message: "error" });

    await expect(getUserSessionElevation()).rejects.toThrow(
        "Failed POST to /user/elevation. Code: 400. Message: error",
    );
});

it("generates user session elevation successfully", async () => {
    const mockRes = { status: 200, data: { status: "OK", data: "generate" } };
    (axios as any).mockResolvedValue(mockRes);
    (hasServiceError as any).mockReturnValue({ errored: false });
    (toData as any).mockReturnValue("generate");

    const result = await generateUserSessionElevation();
    expect(axios).toHaveBeenCalledWith({
        method: "POST",
        url: "/user/elevation",
    });
    expect(result).toBe("generate");
});

it("generates user session elevation with error", async () => {
    const mockRes = { status: 400, data: { status: "KO", message: "error" } };
    (axios as any).mockResolvedValue(mockRes);
    (hasServiceError as any).mockReturnValue({ errored: true, message: "error" });

    await expect(generateUserSessionElevation()).rejects.toThrow(
        "Failed POST to /user/elevation. Code: 400. Message: error",
    );
});

it("verifies user session elevation successfully", async () => {
    const mockRes = { status: 200, data: { status: "OK" } };
    (axios as any).mockResolvedValue(mockRes);

    const result = await verifyUserSessionElevation("otc123");
    expect(axios).toHaveBeenCalledWith({
        method: "PUT",
        url: "/user/elevation",
        data: { otc: "otc123" },
        validateStatus: validateStatusOneTimeCode,
    });
    expect(result).toBe(true);
});

it("verifies user session elevation with error", async () => {
    const mockRes = { status: 400, data: { status: "KO" } };
    (axios as any).mockResolvedValue(mockRes);

    const result = await verifyUserSessionElevation("otc123");
    expect(result).toBe(false);
});

it("deletes user session elevation successfully", async () => {
    const mockRes = { status: 200, data: { status: "OK" } };
    (axios as any).mockResolvedValue(mockRes);

    const result = await deleteUserSessionElevation("delete123");
    expect(axios).toHaveBeenCalledWith({
        method: "DELETE",
        url: "/user/elevation/delete123",
    });
    expect(result).toBe(true);
});

it("deletes user session elevation with error", async () => {
    const mockRes = { status: 400, data: { status: "KO" } };
    (axios as any).mockResolvedValue(mockRes);

    const result = await deleteUserSessionElevation("delete123");
    expect(result).toBe(false);
});
