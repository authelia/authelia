import axios, { AxiosResponse } from "axios";

import {
    OptionalDataServiceResponse,
    ServiceResponse,
    WebauthnAssertionPath,
    WebauthnAttestationPath,
    WebauthnIdentityPathFinish,
} from "@services/Api";
import { SignInResponse } from "@services/SignIn";

export function browserSupportsWebauthn(): boolean {
    return window?.PublicKeyCredential !== undefined && typeof window.PublicKeyCredential === "function";
}

export async function platformAuthenticatorAvailable(): Promise<boolean> {
    if (!browserSupportsWebauthn()) {
        return false;
    }

    return window.PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable();
}

function bufferDecode(value: BufferSource): Uint8Array {
    return Uint8Array.from(atob((value as ArrayBuffer).toString()), (x) => x.charCodeAt(0));
}

function bufferEncode(value: BufferSource): string {
    return btoa(String.fromCharCode.apply(null, Array.from(new Uint8Array(value as ArrayBuffer))))
        .replace(/\+/g, "-")
        .replace(/\//g, "_")
        .replace(/=/g, "");
}

function attestationRequestOptionsDecode(opts: PublicKeyCredentialCreationOptions): PublicKeyCredentialCreationOptions {
    if (opts !== undefined) {
        opts.challenge = bufferDecode(opts.challenge);

        // TODO: Check it this is needed.
        if (opts.user.id != null) {
            opts.user.id = bufferDecode(opts.user.id);
        }
        if (opts.excludeCredentials !== undefined) {
            for (let i = 0; i < opts.excludeCredentials.length; i++) {
                opts.excludeCredentials[i].id = bufferDecode(opts.excludeCredentials[i].id);
            }
        }
    }

    return opts;
}

function assertionRequestOptionsDecode(opts: PublicKeyCredentialRequestOptions): PublicKeyCredentialRequestOptions {
    if (opts !== undefined) {
        opts.challenge = bufferDecode(opts.challenge);

        if (opts.allowCredentials !== undefined) {
            for (let i = 0; i < opts.allowCredentials.length; i++) {
                opts.allowCredentials[i].id = bufferDecode(opts.allowCredentials[i].id);
            }
        }
    }

    return opts;
}

interface AttestationChallengeResponse {
    id: string;
    rawId: string;
    type: string;
    clientExtensionResults: AuthenticationExtensionsClientOutputs;
    transports?: AuthenticatorTransport[];
    response: AttestationResponse;
}

interface AttestationResponse {
    attestationObject: string;
    clientDataJSON: string;
}

interface AssertionChallengeResponse {
    id: string;
    rawId: string;
    type: string;
    clientExtensionResults: AuthenticationExtensionsClientOutputs;
    response: AssertionResponse;
}

interface AssertionResponse {
    authenticatorData: string;
    clientDataJSON: string;
    signature: string;
    userHandle: string;
}

// https://w3c.github.io/webauthn/#sctn-op-make-cred
export enum AttestationResult {
    Success = 1,
    Failure,
    FailureToken,
    FailureExcluded,
    FailureUserConsent,
    FailureUserVerificationOrResidentKey,
    FailureSyntax,
    FailureSupport,
    FailureUnknown,
    FailureWebauthnNotSupported,
}

export enum AssertionResult {
    Success = 1,
    Failure,
    FailureUserConsent,
    FailureU2FFacetID,
    FailureSyntax,
    FailureUnknown,
    FailureUnknownSecurity,
    FailureWebauthnNotSupported,
}

interface AssertionPublicKeyCredential {
    result: AssertionResult;
    credential?: PublicKeyCredential;
}

interface AttestationPublicKeyCredential {
    result: AttestationResult;
    credential?: PublicKeyCredential;
}

interface AuthenticatorAttestationResponseWithGetTransports extends AuthenticatorAttestationResponse {
    getTransports: GetTransports;
}

interface GetTransports {
    (): AuthenticatorTransport[];
}

function createAttestationResponse(credential: PublicKeyCredential): AttestationChallengeResponse {
    const attestationResponse = credential.response as AuthenticatorAttestationResponse;

    let response: AttestationChallengeResponse = {
        id: credential.id,
        rawId: bufferEncode(new Uint8Array(credential.rawId)),
        type: credential.type,
        clientExtensionResults: credential.getClientExtensionResults(),
        response: {
            attestationObject: bufferEncode(new Uint8Array(attestationResponse.attestationObject)),
            clientDataJSON: bufferEncode(new Uint8Array(attestationResponse.clientDataJSON)),
        },
    };

    const attestationResponseWithTransports = attestationResponse as AuthenticatorAttestationResponseWithGetTransports;
    if (
        attestationResponseWithTransports?.getTransports !== undefined &&
        typeof attestationResponseWithTransports.getTransports === "function"
    ) {
        response.transports = attestationResponseWithTransports.getTransports();
    }

    return response;
}

function createAssertionResponse(credential: PublicKeyCredential): AssertionChallengeResponse {
    const assertionResponse = credential.response as AuthenticatorAssertionResponse;

    let userHandle: string;

    if (assertionResponse.userHandle == null) {
        userHandle = "";
    } else {
        userHandle = bufferEncode(new Uint8Array(assertionResponse.userHandle));
    }

    return {
        id: credential.id,
        rawId: bufferEncode(new Uint8Array(credential.rawId)),
        type: credential.type,
        clientExtensionResults: credential.getClientExtensionResults(),
        response: {
            authenticatorData: bufferEncode(new Uint8Array(assertionResponse.authenticatorData)),
            clientDataJSON: bufferEncode(new Uint8Array(assertionResponse.clientDataJSON)),
            signature: bufferEncode(new Uint8Array(assertionResponse.signature)),
            userHandle: userHandle,
        },
    };
}

function assertionDOMExceptionToAssertionResult(
    exception: DOMException,
    requestOptions: PublicKeyCredentialRequestOptions,
): AssertionResult {
    // Docs for this section:
    // https://w3c.github.io/webauthn/#sctn-op-get-assertion
    switch (exception.name) {
        case "UnknownError":
            // § 6.3.3 Step 1 and Step 12.
            return AssertionResult.FailureSyntax;
        case "NotAllowedError":
            // § 6.3.3 Step 6 and Step 7.
            return AssertionResult.FailureUserConsent;
        case "SecurityError":
            // § 10.1 and 10.2 Step 3.
            if (requestOptions.extensions?.appid !== undefined) {
                return AssertionResult.FailureU2FFacetID;
            } else {
                return AssertionResult.FailureUnknownSecurity;
            }
        default:
            console.error(`Unhandled DOMException occurred during WebAuthN assertion: ${exception}`);
            return AssertionResult.FailureUnknown;
    }
}

function attestationDOMExceptionToAttestationResult(exception: DOMException): AttestationResult {
    // Docs for this section:
    // https://w3c.github.io/webauthn/#sctn-op-make-cred
    switch (exception.name) {
        case "UnknownError":
            // § 6.3.2 Step 1 and Step 8.
            return AttestationResult.FailureSyntax;
        case "NotSupportedError":
            // § 6.3.2 Step 2.
            return AttestationResult.FailureSupport;
        case "InvalidStateError":
            // § 6.3.2 Step 3.
            return AttestationResult.FailureExcluded;
        case "NotAllowedError":
            // § 6.3.2 Step 3 and Step 6.
            return AttestationResult.FailureUserConsent;
        // § 6.3.2 Step 4 and Step 5.
        case "ConstraintError":
            return AttestationResult.FailureUserVerificationOrResidentKey;
        default:
            console.error(`Unhandled DOMException occurred during WebAuthN attestation: ${exception}`);
            return AttestationResult.FailureUnknown;
    }
}

async function getAttestationPublicKeyCredential(
    creationOptions: PublicKeyCredentialCreationOptions,
): Promise<AttestationPublicKeyCredential> {
    const result: AttestationPublicKeyCredential = {
        result: AttestationResult.Success,
    };

    try {
        const credential = (await navigator.credentials.create({ publicKey: creationOptions })) as PublicKeyCredential;

        if (credential === undefined || credential === null) {
            result.result = AttestationResult.Failure;
        } else {
            result.result = AttestationResult.Success;
            result.credential = credential;
        }
    } catch (e) {
        result.result = AttestationResult.Failure;

        const exception = e as DOMException;
        if (exception !== undefined) {
            result.result = attestationDOMExceptionToAttestationResult(exception);
        } else {
            console.error(`Unhandled exception occurred during WebAuthN attestation: ${e}`);
        }
    }

    return result;
}

export async function performWebauthnAttestationCeremony(processToken: string): Promise<AttestationResult> {
    if (!browserSupportsWebauthn()) {
        return AttestationResult.FailureWebauthnNotSupported;
    }

    const challengeOptions = await getAttestationChallenge(processToken);

    if (challengeOptions.options === undefined) {
        if (challengeOptions.status === 403) {
            return AttestationResult.FailureToken;
        }
        return AttestationResult.Failure;
    }

    const result = await getAttestationPublicKeyCredential(challengeOptions.options);

    if (result.result !== AttestationResult.Success) {
        return result.result;
    } else if (result.credential === undefined) {
        return AttestationResult.Failure;
    }

    const response = await completeAttestationChallenge(result.credential);

    if (response.data.status === "OK" && response.status === 201) {
        return AttestationResult.Success;
    }

    return AttestationResult.Failure;
}

interface CreationOptions {
    options?: PublicKeyCredentialCreationOptions;
    status: number;
}

async function getAttestationChallenge(processToken: string): Promise<CreationOptions> {
    const response = await axios.post<ServiceResponse<CredentialCreationOptions>>(WebauthnIdentityPathFinish, {
        token: processToken,
    });

    if (response.data.status !== "OK" || response.data.data.publicKey === undefined) {
        return {
            status: response.status,
        };
    }

    return {
        options: attestationRequestOptionsDecode(response.data.data.publicKey),
        status: response.status,
    };
}

async function completeAttestationChallenge(
    credential: PublicKeyCredential,
): Promise<AxiosResponse<OptionalDataServiceResponse<any>>> {
    const attestationResponse = createAttestationResponse(credential);

    return axios.post<OptionalDataServiceResponse<any>>(WebauthnAttestationPath, attestationResponse);
}

export async function performWebauthnAssertionCeremony(): Promise<AssertionResult> {
    if (!browserSupportsWebauthn()) {
        return AssertionResult.FailureWebauthnNotSupported;
    }

    const challengeOptions = await getWebauthnAssertionChallenge();

    if (challengeOptions === undefined) {
        return AssertionResult.Failure;
    }

    const result = await getWebauthnAssertionPublicKeyCredential(challengeOptions);

    if (result.result !== AssertionResult.Success) {
        return result.result;
    } else if (result.credential === undefined) {
        return AssertionResult.Failure;
    }

    const response = await postWebauthnAssertionChallengeResponse(result.credential);

    if (response.data.status === "OK") {
        return AssertionResult.Success;
    }

    return AssertionResult.Failure;
}

export async function getWebauthnAssertionChallenge(): Promise<PublicKeyCredentialRequestOptions | undefined> {
    const response = await axios.get<ServiceResponse<CredentialRequestOptions>>(WebauthnAssertionPath);

    if (response.data.status !== "OK" || response.data.data.publicKey === undefined) {
        return undefined;
    }

    return assertionRequestOptionsDecode(response.data.data.publicKey);
}

export async function postWebauthnAssertionChallengeResponse(
    credential: PublicKeyCredential,
): Promise<AxiosResponse<ServiceResponse<SignInResponse>>> {
    const assertionResponse = createAssertionResponse(credential);

    return axios.post<ServiceResponse<SignInResponse>>(WebauthnAssertionPath, assertionResponse);
}

export async function getWebauthnAssertionPublicKeyCredential(
    requestOptions: PublicKeyCredentialRequestOptions,
): Promise<AssertionPublicKeyCredential> {
    if (!browserSupportsWebauthn()) {
        return {
            result: AssertionResult.FailureWebauthnNotSupported,
        };
    }

    const result: AssertionPublicKeyCredential = {
        result: AssertionResult.Success,
    };

    try {
        const credential = (await navigator.credentials.get({ publicKey: requestOptions })) as PublicKeyCredential;

        if (credential === undefined) {
            result.result = AssertionResult.Failure;
        } else {
            result.result = AssertionResult.Success;
            result.credential = credential;
        }
    } catch (e) {
        result.result = AssertionResult.Failure;

        const exception = e as DOMException;
        if (exception !== undefined) {
            result.result = assertionDOMExceptionToAssertionResult(exception, requestOptions);
        } else {
            console.error(`Unhandled exception occurred during WebAuthN assertion: ${e}`);
        }
    }

    return result;
}
