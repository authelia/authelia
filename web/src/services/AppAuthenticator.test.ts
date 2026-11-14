// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import axios from "axios";

import {
    appAuthenticatorAttestationStatus,
    cancelAppAuthenticatorRegistration,
    deleteAppAuthenticatorDevice,
    getAppAuthenticatorDevices,
    getAppAuthenticatorRegistration,
    getAppAuthenticatorSignInStatus,
    initiateAppAuthenticatorSignIn,
    startAppAuthenticatorRegistration,
    updateAppAuthenticatorDevice,
} from "@services/AppAuthenticator";
import {
    DeleteWithOptionalResponse,
    Get,
    GetWithOptionalData,
    Post,
    PostWithOptionalResponse,
    Put,
} from "@services/Client";

vi.mock("axios", () => ({
    default: vi.fn(),
}));

vi.mock("@services/Api", () => ({
    AppAuthenticatorDevicesPath: "/app/devices",
    AppAuthenticatorRegistrationPath: "/app/registration",
    AppAuthenticatorSignInPath: "/app",
    AppAuthenticatorSignInStatusPath: "/app/status",
    validateStatusAuthentication: vi.fn(),
}));

vi.mock("@services/Client", () => ({
    DeleteWithOptionalResponse: vi.fn(),
    Get: vi.fn(),
    GetWithOptionalData: vi.fn(),
    Post: vi.fn(),
    PostWithOptionalResponse: vi.fn(),
    Put: vi.fn(),
}));

const body = { flow: "flowtype", flowID: "flow", subflow: "sub", targetURL: "url", userCode: "code" };

beforeEach(() => {
    vi.clearAllMocks();
});

it("initiates a sign in challenge", async () => {
    vi.mocked(Post).mockResolvedValue({ expires_at: 1, nonce_length: 2, number: "07" });

    const result = await initiateAppAuthenticatorSignIn(body);

    expect(Post).toHaveBeenCalledWith("/app", body, undefined);
    expect(result).toEqual({ expires_at: 1, nonce_length: 2, number: "07" });
});

it("polls the sign in status with the sign in body", async () => {
    const signal = new AbortController().signal;
    vi.mocked(PostWithOptionalResponse).mockResolvedValue({ state: "pending" });

    const result = await getAppAuthenticatorSignInStatus(body, signal);

    expect(PostWithOptionalResponse).toHaveBeenCalledWith("/app/status", body, signal);
    expect(result).toEqual({ state: "pending" });
});

it("starts a registration", async () => {
    vi.mocked(Put).mockResolvedValue({ expires_at: 1, qr: "authelia-app://x" });

    const result = await startAppAuthenticatorRegistration();

    expect(Put).toHaveBeenCalledWith("/app/registration", undefined, undefined);
    expect(result).toEqual({ expires_at: 1, qr: "authelia-app://x" });
});

it("gets the registration state", async () => {
    vi.mocked(Get).mockResolvedValue({ expires_at: 1, state: "waiting" });

    const result = await getAppAuthenticatorRegistration();

    expect(Get).toHaveBeenCalledWith("/app/registration", undefined);
    expect(result).toEqual({ expires_at: 1, state: "waiting" });
});

it("cancels a registration", async () => {
    vi.mocked(DeleteWithOptionalResponse).mockResolvedValue(undefined);

    await cancelAppAuthenticatorRegistration();

    expect(DeleteWithOptionalResponse).toHaveBeenCalledWith("/app/registration", undefined, undefined);
});

it("returns the devices", async () => {
    const devices = [{ id: 1, name: "Phone" }];
    vi.mocked(GetWithOptionalData).mockResolvedValue(devices);

    const result = await getAppAuthenticatorDevices();

    expect(GetWithOptionalData).toHaveBeenCalledWith("/app/devices");
    expect(result).toEqual(devices);
});

it("returns an empty device list for null data", async () => {
    vi.mocked(GetWithOptionalData).mockResolvedValue(null);

    expect(await getAppAuthenticatorDevices()).toEqual([]);
});

it("renames a device", async () => {
    vi.mocked(axios).mockResolvedValue({ data: { status: "OK" } } as any);

    const result = await updateAppAuthenticatorDevice(4, "Work phone");

    expect(axios).toHaveBeenCalledWith(
        expect.objectContaining({ data: { name: "Work phone" }, method: "PUT", url: "/app/devices/4" }),
    );
    expect(result.data.status).toBe("OK");
});

it("deletes a device", async () => {
    vi.mocked(axios).mockResolvedValue({ data: { status: "OK" } } as any);

    await deleteAppAuthenticatorDevice(4);

    expect(axios).toHaveBeenCalledWith(expect.objectContaining({ method: "DELETE", url: "/app/devices/4" }));
});

describe("attestation status", () => {
    it("reports not attested for format none", () => {
        expect(appAuthenticatorAttestationStatus({ attestation_format: "none" })).toEqual({ kind: "none" });
    });

    it("reports not attested without an identity", () => {
        expect(appAuthenticatorAttestationStatus({ attestation_format: "android-key" })).toEqual({ kind: "none" });
    });

    it("reports the official android app", () => {
        expect(
            appAuthenticatorAttestationStatus({
                attestation_format: "android-key",
                attestation_identity: "com.authelia.mobile",
            }),
        ).toEqual({ kind: "official" });
    });

    it("reports the official ios app", () => {
        expect(
            appAuthenticatorAttestationStatus({
                attestation_format: "apple-appattest",
                attestation_identity: "ABCDE12345.com.authelia.mobile",
            }),
        ).toEqual({ kind: "official" });
    });

    it("reports other identities as verified", () => {
        expect(
            appAuthenticatorAttestationStatus({
                attestation_format: "apple-appattest",
                attestation_identity: "ABCDE12345.com.example.app",
            }),
        ).toEqual({ identity: "ABCDE12345.com.example.app", kind: "verified" });
    });

    it("does not treat a lookalike ios bundle as official", () => {
        expect(
            appAuthenticatorAttestationStatus({
                attestation_format: "apple-appattest",
                attestation_identity: "ABCDE12345.evil.com.authelia.mobile",
            }),
        ).toEqual({ identity: "ABCDE12345.evil.com.authelia.mobile", kind: "verified" });
    });

    it("does not treat a lookalike android package as official", () => {
        expect(
            appAuthenticatorAttestationStatus({
                attestation_format: "android-key",
                attestation_identity: "evil.com.authelia.mobile",
            }),
        ).toEqual({ identity: "evil.com.authelia.mobile", kind: "verified" });
    });
});
