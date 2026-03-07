import axios from "axios";

import { TOTPAlgorithm } from "@models/TOTPConfiguration";
import { validateStatusAuthentication } from "@services/Api";
import { Get } from "@services/Client";
import {
    deleteUserTOTPConfiguration,
    getTOTPOptions,
    getUserInfoTOTPConfiguration,
    getUserInfoTOTPConfigurationOptional,
} from "@services/UserInfoTOTPConfiguration";

vi.mock("axios");
vi.mock("@services/Api", () => ({
    CompleteTOTPSignInPath: "/totp/signin",
    TOTPConfigurationPath: "/totp/config",
    TOTPRegistrationPath: "/totp/register",
    validateStatusAuthentication: vi.fn(),
}));
vi.mock("@services/Client", () => ({
    Get: vi.fn(),
}));

it("gets user info TOTP configuration", async () => {
    (Get as any).mockResolvedValue({
        algorithm: "SHA1",
        created_at: "2024-01-01T00:00:00Z",
        digits: 6,
        issuer: "Authelia",
        last_used_at: "2024-06-01T00:00:00Z",
        period: 30,
    });

    const result = await getUserInfoTOTPConfiguration();
    expect(Get).toHaveBeenCalledWith("/totp/config");
    expect(result.algorithm).toBe(TOTPAlgorithm.SHA1);
    expect(result.created_at).toBeInstanceOf(Date);
    expect(result.last_used_at).toBeInstanceOf(Date);
    expect(result.digits).toBe(6);
    expect(result.period).toBe(30);
});

it("gets user info TOTP configuration without last_used_at", async () => {
    (Get as any).mockResolvedValue({
        algorithm: "SHA256",
        created_at: "2024-01-01T00:00:00Z",
        digits: 8,
        issuer: "Authelia",
        period: 30,
    });

    const result = await getUserInfoTOTPConfiguration();
    expect(result.algorithm).toBe(TOTPAlgorithm.SHA256);
    expect(result.last_used_at).toBeUndefined();
});

it("gets optional TOTP configuration when present", async () => {
    (axios.get as any).mockResolvedValue({
        data: {
            data: {
                algorithm: "SHA512",
                created_at: "2024-01-01T00:00:00Z",
                digits: 6,
                issuer: "Authelia",
                period: 30,
            },
            status: "OK",
        },
        status: 200,
    });

    const result = await getUserInfoTOTPConfigurationOptional();
    expect(result).not.toBeNull();
    expect(result!.algorithm).toBe(TOTPAlgorithm.SHA512);

    const validateStatus = (axios.get as any).mock.calls[0][1].validateStatus;
    expect(validateStatus(200)).toBe(true);
    expect(validateStatus(404)).toBe(true);
    expect(validateStatus(500)).toBe(false);
});

it("returns null for optional TOTP configuration when 404", async () => {
    (axios.get as any).mockResolvedValue({
        data: { status: "KO" },
        status: 404,
    });

    const result = await getUserInfoTOTPConfigurationOptional();
    expect(result).toBeNull();

    const validateStatus = (axios.get as any).mock.calls[0][1].validateStatus;
    expect(validateStatus(404)).toBe(true);
});

it("returns null for optional TOTP configuration when KO", async () => {
    (axios.get as any).mockResolvedValue({
        data: { status: "KO" },
        status: 200,
    });

    const result = await getUserInfoTOTPConfigurationOptional();
    expect(result).toBeNull();
});

it("gets TOTP options", async () => {
    (Get as any).mockResolvedValue({
        algorithm: "SHA1",
        algorithms: ["SHA1", "SHA256", "SHA512"],
        length: 6,
        lengths: [6, 8],
        period: 30,
        periods: [30, 60],
    });

    const result = await getTOTPOptions();
    expect(Get).toHaveBeenCalledWith("/totp/register");
    expect(result.algorithm).toBe(TOTPAlgorithm.SHA1);
    expect(result.algorithms).toEqual([TOTPAlgorithm.SHA1, TOTPAlgorithm.SHA256, TOTPAlgorithm.SHA512]);
    expect(result.length).toBe(6);
    expect(result.periods).toEqual([30, 60]);
});

it("deletes user TOTP configuration", async () => {
    (axios as any).mockResolvedValue({ data: { status: "OK" }, status: 200 });
    await deleteUserTOTPConfiguration();
    expect(axios).toHaveBeenCalledWith({
        method: "DELETE",
        url: "/totp/signin",
        validateStatus: validateStatusAuthentication,
    });
});
