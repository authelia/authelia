import axios from "axios";

import { hasServiceError, toData, validateStatusOneTimeCode } from "@services/Api";
import {
    deleteUserSessionElevation,
    generateUserSessionElevation,
    getUserSessionElevation,
    verifyUserSessionElevation,
} from "@services/UserSessionElevation";

vi.mock("axios");
vi.mock("@services/Api", () => ({
    hasServiceError: vi.fn(),
    toData: vi.fn(),
    UserSessionElevationPath: "/user/elevation",
    validateStatusOneTimeCode: vi.fn(),
}));

it("gets user session elevation successfully", async () => {
    const mockRes = { data: { data: "elevation", status: "OK" }, status: 200 };
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
    const mockRes = { data: { message: "error", status: "KO" }, status: 400 };
    (axios as any).mockResolvedValue(mockRes);
    (hasServiceError as any).mockReturnValue({ errored: true, message: "error" });

    await expect(getUserSessionElevation()).rejects.toThrow(
        "Failed POST to /user/elevation. Code: 400. Message: error",
    );
});

it("generates user session elevation successfully", async () => {
    const mockRes = { data: { data: "generate", status: "OK" }, status: 200 };
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
    const mockRes = { data: { message: "error", status: "KO" }, status: 400 };
    (axios as any).mockResolvedValue(mockRes);
    (hasServiceError as any).mockReturnValue({ errored: true, message: "error" });

    await expect(generateUserSessionElevation()).rejects.toThrow(
        "Failed POST to /user/elevation. Code: 400. Message: error",
    );
});

it("verifies user session elevation successfully", async () => {
    const mockRes = { data: { status: "OK" }, status: 200 };
    (axios as any).mockResolvedValue(mockRes);

    const result = await verifyUserSessionElevation("otc123");
    expect(axios).toHaveBeenCalledWith({
        data: { otc: "otc123" },
        method: "PUT",
        url: "/user/elevation",
        validateStatus: validateStatusOneTimeCode,
    });
    expect(result).toBe(true);
});

it("verifies user session elevation with error", async () => {
    const mockRes = { data: { status: "KO" }, status: 400 };
    (axios as any).mockResolvedValue(mockRes);

    const result = await verifyUserSessionElevation("otc123");
    expect(result).toBe(false);
});

it("deletes user session elevation successfully", async () => {
    const mockRes = { data: { status: "OK" }, status: 200 };
    (axios as any).mockResolvedValue(mockRes);

    const result = await deleteUserSessionElevation("delete123");
    expect(axios).toHaveBeenCalledWith({
        method: "DELETE",
        url: "/user/elevation/delete123",
    });
    expect(result).toBe(true);
});

it("deletes user session elevation with error", async () => {
    const mockRes = { data: { status: "KO" }, status: 400 };
    (axios as any).mockResolvedValue(mockRes);

    const result = await deleteUserSessionElevation("delete123");
    expect(result).toBe(false);
});
