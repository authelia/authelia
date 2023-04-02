// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import axios, { AxiosResponse } from "axios";

import {
    AssertionPublicKeyCredentialResult,
    AssertionResult,
    AttestationPublicKeyCredential,
    AttestationPublicKeyCredentialJSON,
    AttestationPublicKeyCredentialResult,
    AttestationResult,
    AuthenticatorAttestationResponseFuture,
    CredentialCreation,
    CredentialRequest,
    PublicKeyCredentialCreationOptionsJSON,
    PublicKeyCredentialCreationOptionsStatus,
    PublicKeyCredentialDescriptorJSON,
    PublicKeyCredentialJSON,
    PublicKeyCredentialRequestOptionsJSON,
    PublicKeyCredentialRequestOptionsStatus,
} from "@models/Webauthn";
import {
    OptionalDataServiceResponse,
    ServiceResponse,
    WebauthnAssertionPath,
    WebauthnAttestationPath,
    WebauthnIdentityFinishPath,
} from "@services/Api";
import { SignInResponse } from "@services/SignIn";
import { getBase64WebEncodingFromBytes, getBytesFromBase64 } from "@utils/Base64";

export function isWebauthnSecure(): boolean {
    if (window.isSecureContext) {
        return true;
    }

    return window.location.hostname === "localhost" || window.location.hostname === "127.0.0.1";
}

export function isWebauthnSupported(): boolean {
    return window?.PublicKeyCredential !== undefined && typeof window.PublicKeyCredential === "function";
}

export async function isWebauthnPlatformAuthenticatorAvailable(): Promise<boolean> {
    if (!isWebauthnSupported()) {
        return false;
    }

    return window.PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable();
}

function arrayBufferEncode(value: ArrayBuffer): string {
    return getBase64WebEncodingFromBytes(new Uint8Array(value));
}

function arrayBufferDecode(value: string): ArrayBuffer {
    return getBytesFromBase64(value);
}

function decodePublicKeyCredentialDescriptor(
    descriptor: PublicKeyCredentialDescriptorJSON,
): PublicKeyCredentialDescriptor {
    return {
        id: arrayBufferDecode(descriptor.id),
        type: descriptor.type,
        transports: descriptor.transports,
    };
}

function decodePublicKeyCredentialCreationOptions(
    options: PublicKeyCredentialCreationOptionsJSON,
): PublicKeyCredentialCreationOptions {
    return {
        attestation: options.attestation,
        authenticatorSelection: options.authenticatorSelection,
        challenge: arrayBufferDecode(options.challenge),
        excludeCredentials: options.excludeCredentials?.map(decodePublicKeyCredentialDescriptor),
        extensions: options.extensions,
        pubKeyCredParams: options.pubKeyCredParams,
        rp: options.rp,
        timeout: options.timeout,
        user: {
            displayName: options.user.displayName,
            id: arrayBufferDecode(options.user.id),
            name: options.user.name,
        },
    };
}

function decodePublicKeyCredentialRequestOptions(
    options: PublicKeyCredentialRequestOptionsJSON,
): PublicKeyCredentialRequestOptions {
    let allowCredentials: PublicKeyCredentialDescriptor[] | undefined = undefined;

    if (options.allowCredentials?.length !== 0) {
        allowCredentials = options.allowCredentials?.map(decodePublicKeyCredentialDescriptor);
    }

    return {
        allowCredentials: allowCredentials,
        challenge: arrayBufferDecode(options.challenge),
        extensions: options.extensions,
        rpId: options.rpId,
        timeout: options.timeout,
        userVerification: options.userVerification,
    };
}

function encodeAttestationPublicKeyCredential(
    credential: AttestationPublicKeyCredential,
): AttestationPublicKeyCredentialJSON {
    const response = credential.response as AuthenticatorAttestationResponseFuture;

    let transports: AuthenticatorTransport[] | undefined;

    if (response?.getTransports !== undefined && typeof response.getTransports === "function") {
        transports = response.getTransports();
    }

    return {
        id: credential.id,
        type: credential.type,
        rawId: arrayBufferEncode(credential.rawId),
        clientExtensionResults: credential.getClientExtensionResults(),
        response: {
            attestationObject: arrayBufferEncode(response.attestationObject),
            clientDataJSON: arrayBufferEncode(response.clientDataJSON),
        },
        transports: transports,
    };
}

function encodeAssertionPublicKeyCredential(
    credential: PublicKeyCredential,
    targetURL: string | undefined,
    workflow: string | undefined,
    workflowID: string | undefined,
): PublicKeyCredentialJSON {
    const response = credential.response as AuthenticatorAssertionResponse;

    let userHandle: string;

    if (response.userHandle == null) {
        userHandle = "";
    } else {
        userHandle = arrayBufferEncode(response.userHandle);
    }

    return {
        id: credential.id,
        type: credential.type,
        rawId: arrayBufferEncode(credential.rawId),
        clientExtensionResults: credential.getClientExtensionResults(),
        response: {
            authenticatorData: arrayBufferEncode(response.authenticatorData),
            clientDataJSON: arrayBufferEncode(response.clientDataJSON),
            signature: arrayBufferEncode(response.signature),
            userHandle: userHandle,
        },
        targetURL: targetURL,
        workflow: workflow,
        workflowID: workflowID,
    };
}

function getAttestationResultFromDOMException(exception: DOMException): AttestationResult {
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
        case "ConstraintError":
            // § 6.3.2 Step 4.
            return AttestationResult.FailureUserVerificationOrResidentKey;
        default:
            console.error(`Unhandled DOMException occurred during WebAuthN attestation: ${exception}`);
            return AttestationResult.FailureUnknown;
    }
}

function getAssertionResultFromDOMException(
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
            if (requestOptions.extensions?.appid !== undefined) {
                // § 10.1 and 10.2 Step 3.
                return AssertionResult.FailureU2FFacetID;
            } else {
                return AssertionResult.FailureUnknownSecurity;
            }
        default:
            console.error(`Unhandled DOMException occurred during WebAuthN assertion: ${exception}`);
            return AssertionResult.FailureUnknown;
    }
}

async function getAttestationCreationOptions(token: string): Promise<PublicKeyCredentialCreationOptionsStatus> {
    let response: AxiosResponse<ServiceResponse<CredentialCreation>>;

    response = await axios.post<ServiceResponse<CredentialCreation>>(WebauthnIdentityFinishPath, {
        token: token,
    });

    if (response.data.status !== "OK" || response.data.data == null) {
        return {
            status: response.status,
        };
    }

    return {
        options: decodePublicKeyCredentialCreationOptions(response.data.data.publicKey),
        status: response.status,
    };
}

export async function getAssertionRequestOptions(): Promise<PublicKeyCredentialRequestOptionsStatus> {
    let response: AxiosResponse<ServiceResponse<CredentialRequest>>;

    response = await axios.get<ServiceResponse<CredentialRequest>>(WebauthnAssertionPath);

    if (response.data.status !== "OK" || response.data.data == null) {
        return {
            status: response.status,
        };
    }

    return {
        options: decodePublicKeyCredentialRequestOptions(response.data.data.publicKey),
        status: response.status,
    };
}

async function getAttestationPublicKeyCredentialResult(
    creationOptions: PublicKeyCredentialCreationOptions,
): Promise<AttestationPublicKeyCredentialResult> {
    const result: AttestationPublicKeyCredentialResult = {
        result: AttestationResult.Success,
    };

    try {
        result.credential = (await navigator.credentials.create({
            publicKey: creationOptions,
        })) as AttestationPublicKeyCredential;
    } catch (e) {
        result.result = AttestationResult.Failure;

        const exception = e as DOMException;
        if (exception !== undefined) {
            result.result = getAttestationResultFromDOMException(exception);

            return result;
        } else {
            console.error(`Unhandled exception occurred during WebAuthN attestation: ${e}`);
        }
    }

    if (result.credential == null) {
        result.result = AttestationResult.Failure;
    } else {
        result.result = AttestationResult.Success;
    }

    return result;
}

export async function getAssertionPublicKeyCredentialResult(
    requestOptions: PublicKeyCredentialRequestOptions,
): Promise<AssertionPublicKeyCredentialResult> {
    const result: AssertionPublicKeyCredentialResult = {
        result: AssertionResult.Success,
    };

    try {
        result.credential = (await navigator.credentials.get({ publicKey: requestOptions })) as PublicKeyCredential;
    } catch (e) {
        result.result = AssertionResult.Failure;

        const exception = e as DOMException;
        if (exception !== undefined) {
            result.result = getAssertionResultFromDOMException(exception, requestOptions);

            return result;
        } else {
            console.error(`Unhandled exception occurred during WebAuthN assertion: ${e}`);
        }
    }

    if (result.credential == null) {
        result.result = AssertionResult.Failure;
    } else {
        result.result = AssertionResult.Success;
    }

    return result;
}

async function postAttestationPublicKeyCredentialResult(
    credential: AttestationPublicKeyCredential,
): Promise<AxiosResponse<OptionalDataServiceResponse<any>>> {
    const credentialJSON = encodeAttestationPublicKeyCredential(credential);

    return axios.post<OptionalDataServiceResponse<any>>(WebauthnAttestationPath, credentialJSON);
}

export async function postAssertionPublicKeyCredentialResult(
    credential: PublicKeyCredential,
    targetURL: string | undefined,
    workflow?: string,
    workflowID?: string,
): Promise<AxiosResponse<ServiceResponse<SignInResponse>>> {
    const credentialJSON = encodeAssertionPublicKeyCredential(credential, targetURL, workflow, workflowID);

    return axios.post<ServiceResponse<SignInResponse>>(WebauthnAssertionPath, credentialJSON);
}

export async function performAttestationCeremony(token: string): Promise<AttestationResult> {
    const attestationCreationOpts = await getAttestationCreationOptions(token);

    if (attestationCreationOpts.status !== 200 || attestationCreationOpts.options == null) {
        if (attestationCreationOpts.status === 403) {
            return AttestationResult.FailureToken;
        }

        return AttestationResult.Failure;
    }

    const attestationResult = await getAttestationPublicKeyCredentialResult(attestationCreationOpts.options);

    if (attestationResult.result !== AttestationResult.Success) {
        return attestationResult.result;
    } else if (attestationResult.credential == null) {
        return AttestationResult.Failure;
    }

    const response = await postAttestationPublicKeyCredentialResult(attestationResult.credential);

    if (response.data.status === "OK" && (response.status === 200 || response.status === 201)) {
        return AttestationResult.Success;
    }

    return AttestationResult.Failure;
}

export async function performAssertionCeremony(
    targetURL?: string,
    workflow?: string,
    workflowID?: string,
): Promise<AssertionResult> {
    const assertionRequestOpts = await getAssertionRequestOptions();

    if (assertionRequestOpts.status !== 200 || assertionRequestOpts.options == null) {
        return AssertionResult.FailureChallenge;
    }

    const assertionResult = await getAssertionPublicKeyCredentialResult(assertionRequestOpts.options);

    if (assertionResult.result !== AssertionResult.Success) {
        return assertionResult.result;
    } else if (assertionResult.credential == null) {
        return AssertionResult.Failure;
    }

    const response = await postAssertionPublicKeyCredentialResult(
        assertionResult.credential,
        targetURL,
        workflow,
        workflowID,
    );

    if (response.data.status === "OK" && response.status === 200) {
        return AssertionResult.Success;
    }

    return AssertionResult.Failure;
}
