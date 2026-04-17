import { startAuthentication, startRegistration } from "@simplewebauthn/browser";
import axios from "axios";

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
beforeEach(() => {
    vi.spyOn(console, "error").mockImplementation(() => {});
});

vi.mock("@services/Api", () => ({
    FirstFactorPasskeyPath: "/firstfactor/passkey",
    validateStatusAuthentication: vi.fn(),
    validateStatusWebAuthnCreation: vi.fn(),
    WebAuthnAssertionPath: "/webauthn/assertion",
    WebAuthnCredentialPath: "/webauthn/credential",
    WebAuthnRegistrationPath: "/webauthn/registration",
}));

it("handles successful webauthn options", async () => {
    const mockRes = { data: { data: { publicKey: "options" }, status: "OK" }, status: 200 };
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

it("handles webauthn result with AbortError dom exception", async () => {
    (startAuthentication as any).mockRejectedValue(new DOMException("test", "AbortError"));

    const result = await getWebAuthnResult({} as any);
    expect(result.result).toBe(AssertionResult.FailureUserConsent);
});

it("handles webauthn result with UnknownError dom exception", async () => {
    (startAuthentication as any).mockRejectedValue(new DOMException("test", "UnknownError"));

    const result = await getWebAuthnResult({} as any);
    expect(result.result).toBe(AssertionResult.FailureSyntax);
});

it("handles webauthn result with InvalidStateError dom exception", async () => {
    (startAuthentication as any).mockRejectedValue(new DOMException("test", "InvalidStateError"));

    const result = await getWebAuthnResult({} as any);
    expect(result.result).toBe(AssertionResult.FailureUnrecognized);
});

it("handles webauthn result with SecurityError and appid extension", async () => {
    (startAuthentication as any).mockRejectedValue(new DOMException("test", "SecurityError"));

    const result = await getWebAuthnResult({ extensions: { appid: "test" } } as any);
    expect(result.result).toBe(AssertionResult.FailureU2FFacetID);
});

it("handles webauthn result with SecurityError without appid extension", async () => {
    (startAuthentication as any).mockRejectedValue(new DOMException("test", "SecurityError"));

    const result = await getWebAuthnResult({} as any);
    expect(result.result).toBe(AssertionResult.FailureUnknownSecurity);
});

it("handles webauthn result with unhandled dom exception", async () => {
    (startAuthentication as any).mockRejectedValue(new DOMException("test", "DataError"));

    const result = await getWebAuthnResult({} as any);
    expect(result.result).toBe(AssertionResult.FailureUnknown);
});

it("handles webauthn result with null response", async () => {
    (startAuthentication as any).mockResolvedValue(null);

    const result = await getWebAuthnResult({} as any);
    expect(result.result).toBe(AssertionResult.Failure);
});

it("handles webauthn response posting", async () => {
    (axios.post as any).mockResolvedValue("response");

    const result = await postWebAuthnResponse("authResponse" as any, "url", "flow", "flowtype", "sub", "code");
    expect(axios.post).toHaveBeenCalledWith(
        "/webauthn/assertion",
        {
            flow: "flowtype",
            flowID: "flow",
            response: "authResponse",
            subflow: "sub",
            targetURL: "url",
            userCode: "code",
        },
        { signal: undefined },
    );
    expect(result).toBe("response");
});

it("handles successful webauthn passkey options", async () => {
    const mockRes = { data: { data: { publicKey: "options" }, status: "OK" }, status: 200 };
    (axios.get as any).mockResolvedValue(mockRes);

    const result = await getWebAuthnPasskeyOptions();
    expect(result).toEqual({ options: "options", status: 200 });
});

it("handles webauthn passkey options with no data", async () => {
    const mockRes = { data: { status: "KO" }, status: 400 };
    (axios.get as any).mockResolvedValue(mockRes);

    const result = await getWebAuthnPasskeyOptions();
    expect(result).toEqual({ status: 400 });
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
    expect(axios.post).toHaveBeenCalledWith(
        "/firstfactor/passkey",
        {
            flow: "flowtype",
            flowID: "flow",
            keepMeLoggedIn: true,
            requestMethod: "POST",
            response: "authResponse",
            subflow: "sub",
            targetURL: "url",
        },
        { signal: undefined },
    );
    expect(result).toBe("response");
});

it("handles successful webauthn registration options", async () => {
    const mockRes = { data: { data: { publicKey: "options" }, status: "OK" }, status: 200 };
    (axios.put as any).mockResolvedValue(mockRes);

    const result = await getWebAuthnRegistrationOptions("desc");
    expect(axios.put).toHaveBeenCalledWith(
        "/webauthn/registration",
        { description: "desc" },
        { validateStatus: expect.any(Function) },
    );
    expect(result).toEqual({ options: "options", status: 200 });
});

it("handles webauthn registration options with no data", async () => {
    const mockRes = { data: { status: "KO" }, status: 400 };
    (axios.put as any).mockResolvedValue(mockRes);

    const result = await getWebAuthnRegistrationOptions("desc");
    expect(result).toEqual({ status: 400 });
});

it("handles successful webauthn registration start", async () => {
    (startRegistration as any).mockResolvedValue("response");

    const result = await startWebAuthnRegistration({} as any);
    expect(result.result).toBe(AttestationResult.Success);
    expect(result.response).toBe("response");
});

it("handles webauthn registration start with AbortError dom exception", async () => {
    (startRegistration as any).mockRejectedValue(new DOMException("test", "AbortError"));

    const result = await startWebAuthnRegistration({} as any);
    expect(result.result).toBe(AttestationResult.FailureUserConsent);
});

it("handles webauthn registration start with UnknownError dom exception", async () => {
    (startRegistration as any).mockRejectedValue(new DOMException("test", "UnknownError"));

    const result = await startWebAuthnRegistration({} as any);
    expect(result.result).toBe(AttestationResult.FailureSyntax);
});

it("handles webauthn registration start with NotSupportedError dom exception", async () => {
    (startRegistration as any).mockRejectedValue(new DOMException("test", "NotSupportedError"));

    const result = await startWebAuthnRegistration({} as any);
    expect(result.result).toBe(AttestationResult.FailureSupport);
});

it("handles webauthn registration start with InvalidStateError dom exception", async () => {
    (startRegistration as any).mockRejectedValue(new DOMException("test", "InvalidStateError"));

    const result = await startWebAuthnRegistration({} as any);
    expect(result.result).toBe(AttestationResult.FailureExcluded);
});

it("handles webauthn registration start with ConstraintError dom exception", async () => {
    (startRegistration as any).mockRejectedValue(new DOMException("test", "ConstraintError"));

    const result = await startWebAuthnRegistration({} as any);
    expect(result.result).toBe(AttestationResult.FailureUserVerificationOrResidentKey);
});

it("handles webauthn registration start with unhandled dom exception", async () => {
    (startRegistration as any).mockRejectedValue(new DOMException("test", "DataError"));

    const result = await startWebAuthnRegistration({} as any);
    expect(result.result).toBe(AttestationResult.FailureUnknown);
});

it("handles successful webauthn registration finish", async () => {
    const mockRes = { data: { status: "OK" }, status: 200 };
    (axios.post as any).mockResolvedValue(mockRes);

    const result = await finishWebAuthnRegistration({} as any);
    expect(result).toEqual({ message: "", status: AttestationResult.Success });
});

it("handles webauthn registration finish with error", async () => {
    const mockError = { response: {} };
    (axios.post as any).mockRejectedValue(mockError);

    const result = await finishWebAuthnRegistration({} as any);
    expect(result).toEqual({ message: "Device registration failed.", status: AttestationResult.Failure });
});

it("handles webauthn registration finish with axios error message", async () => {
    const { AxiosError } = await import("axios");
    const message = "Device registration failed.";
    const axiosError = new AxiosError("fail", undefined, undefined, undefined, { data: { message } } as any);
    (axios.post as any).mockRejectedValue(axiosError);

    const result = await finishWebAuthnRegistration({} as any);
    expect(result).toEqual({ message, status: AttestationResult.Failure });
});

it("forwards the abort signal through GET, assertion POST and passkey POST", async () => {
    const signal = new AbortController().signal;
    const mockRes = { data: { data: { publicKey: "options" }, status: "OK" }, status: 200 };
    (axios.get as any).mockResolvedValue(mockRes);
    (axios.post as any).mockResolvedValue("response");

    await getWebAuthnOptions(signal);
    expect(axios.get).toHaveBeenLastCalledWith("/webauthn/assertion", { signal });

    await getWebAuthnPasskeyOptions(signal);
    expect(axios.get).toHaveBeenLastCalledWith("/firstfactor/passkey", { signal });

    await postWebAuthnResponse("authResponse" as any, "url", "flow", "flowtype", "sub", "code", signal);
    expect(axios.post).toHaveBeenLastCalledWith("/webauthn/assertion", expect.any(Object), { signal });

    await postWebAuthnPasskeyResponse("authResponse" as any, true, "url", "POST", "flow", "flowtype", "sub", signal);
    expect(axios.post).toHaveBeenLastCalledWith("/firstfactor/passkey", expect.any(Object), { signal });
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
        data: { description: "desc" },
        method: "PUT",
        url: "/webauthn/credential/cred123",
        validateStatus: expect.any(Function),
    });
    expect(result).toBe("response");
});
