import { startAuthentication, startRegistration } from "@simplewebauthn/browser";
import axios from "axios";
import { vi } from "vitest";

import { AssertionResult, AttestationResult } from "@models/WebAuthn";
import {
    deleteUserWebAuthnCredential,
    finishWebAuthnRegistration,
    getWebAuthnOptions,
    getWebAuthnPasskeyOptions,
    getWebAuthnRegistrationOptions,
    getWebAuthnResult,
    postWebAuthnPasskeyResponse,
    postWebAuthnResponse,
    startWebAuthnRegistration,
    updateUserWebAuthnCredential,
} from "@services/WebAuthn";

vi.mock("axios");
vi.mock("@simplewebauthn/browser", () => ({
    startAuthentication: vi.fn(),
    startRegistration: vi.fn(),
}));
vi.mock("@services/Api", () => ({
    WebAuthnAssertionPath: "/webauthn/assertion",
    FirstFactorPasskeyPath: "/firstfactor/passkey",
    WebAuthnRegistrationPath: "/webauthn/registration",
    WebAuthnCredentialPath: "/webauthn/credential",
    validateStatusAuthentication: vi.fn(),
    validateStatusWebAuthnCreation: vi.fn(),
}));

it("handles successful webauthn options", async () => {
    const mockRes = { data: { status: "OK", data: { publicKey: "options" } }, status: 200 };
    (axios.get as any).mockResolvedValue(mockRes);

    const result = await getWebAuthnOptions();
    expect(result).toEqual({ options: "options", status: 200 });
});

it("handles webauthn options with no data", async () => {
    const mockRes = { data: { status: "KO" }, status: 400 };
    (axios.get as any).mockResolvedValue(mockRes);

    const result = await getWebAuthnOptions();
    expect(result).toEqual({ status: 400 });
});

it("handles successful webauthn result", async () => {
    (startAuthentication as any).mockResolvedValue("response");

    const result = await getWebAuthnResult({} as any);
    expect(result.result).toBe(AssertionResult.Success);
    expect(result.response).toBe("response");
});

it("handles webauthn result with dom exception", async () => {
    (startAuthentication as any).mockRejectedValue(new DOMException("test", "AbortError"));

    const result = await getWebAuthnResult({} as any);
    expect(result.result).toBe(AssertionResult.FailureUserConsent);
});

it("handles webauthn result with null response", async () => {
    (startAuthentication as any).mockResolvedValue(null);

    const result = await getWebAuthnResult({} as any);
    expect(result.result).toBe(AssertionResult.Failure);
});

it("handles webauthn response posting", async () => {
    (axios.post as any).mockResolvedValue("response");

    const result = await postWebAuthnResponse("authResponse" as any, "url", "flow", "flowtype", "sub", "code");
    expect(axios.post).toHaveBeenCalledWith("/webauthn/assertion", {
        response: "authResponse",
        targetURL: "url",
        flowID: "flow",
        flow: "flowtype",
        subflow: "sub",
        userCode: "code",
    });
    expect(result).toBe("response");
});

it("handles successful webauthn passkey options", async () => {
    const mockRes = { data: { status: "OK", data: { publicKey: "options" } }, status: 200 };
    (axios.get as any).mockResolvedValue(mockRes);

    const result = await getWebAuthnPasskeyOptions();
    expect(result).toEqual({ options: "options", status: 200 });
});

it("handles webauthn passkey response posting", async () => {
    (axios.post as any).mockResolvedValue("response");

    const result = await postWebAuthnPasskeyResponse(
        "authResponse" as any,
        true,
        "url",
        "POST",
        "flow",
        "flowtype",
        "sub",
    );
    expect(axios.post).toHaveBeenCalledWith("/firstfactor/passkey", {
        response: "authResponse",
        keepMeLoggedIn: true,
        targetURL: "url",
        requestMethod: "POST",
        flowID: "flow",
        flow: "flowtype",
        subflow: "sub",
    });
    expect(result).toBe("response");
});

it("handles successful webauthn registration options", async () => {
    const mockRes = { data: { status: "OK", data: { publicKey: "options" } }, status: 200 };
    (axios.put as any).mockResolvedValue(mockRes);

    const result = await getWebAuthnRegistrationOptions("desc");
    expect(axios.put).toHaveBeenCalledWith(
        "/webauthn/registration",
        { description: "desc" },
        { validateStatus: expect.any(Function) },
    );
    expect(result).toEqual({ options: "options", status: 200 });
});

it("handles successful webauthn registration start", async () => {
    (startRegistration as any).mockResolvedValue("response");

    const result = await startWebAuthnRegistration({} as any);
    expect(result.result).toBe(AttestationResult.Success);
    expect(result.response).toBe("response");
});

it("handles webauthn registration start with dom exception", async () => {
    (startRegistration as any).mockRejectedValue(new DOMException("test", "AbortError"));

    const result = await startWebAuthnRegistration({} as any);
    expect(result.result).toBe(AttestationResult.FailureUserConsent);
});

it("handles successful webauthn registration finish", async () => {
    const mockRes = { data: { status: "OK" }, status: 200 };
    (axios.post as any).mockResolvedValue(mockRes);

    const result = await finishWebAuthnRegistration({} as any);
    expect(result).toEqual({ status: AttestationResult.Success, message: "" });
});

it("handles webauthn registration finish with error", async () => {
    const mockError = { response: {} };
    (axios.post as any).mockRejectedValue(mockError);

    const result = await finishWebAuthnRegistration({} as any);
    expect(result).toEqual({ status: AttestationResult.Failure, message: "Device registration failed." });
});

it("handles webauthn credential deletion", async () => {
    (axios as any).mockResolvedValue("response");

    const result = await deleteUserWebAuthnCredential("cred123");
    expect(axios).toHaveBeenCalledWith({
        method: "DELETE",
        url: "/webauthn/credential/cred123",
        validateStatus: expect.any(Function),
    });
    expect(result).toBe("response");
});

it("handles webauthn credential update", async () => {
    (axios as any).mockResolvedValue("response");

    const result = await updateUserWebAuthnCredential("cred123", "desc");
    expect(axios).toHaveBeenCalledWith({
        method: "PUT",
        url: "/webauthn/credential/cred123",
        data: { description: "desc" },
        validateStatus: expect.any(Function),
    });
    expect(result).toBe("response");
});
