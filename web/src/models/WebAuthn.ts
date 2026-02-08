import {
    AuthenticationResponseJSON,
    PublicKeyCredentialCreationOptionsJSON,
    PublicKeyCredentialRequestOptionsJSON,
    RegistrationResponseJSON,
} from "@simplewebauthn/browser";

export interface PublicKeyCredentialCreationOptionsStatus {
    options?: PublicKeyCredentialCreationOptionsJSON;
    status: number;
}

export interface CredentialCreation {
    publicKey: PublicKeyCredentialCreationOptionsJSON;
}

export interface PublicKeyCredentialRequestOptionsStatus {
    options?: PublicKeyCredentialRequestOptionsJSON;
    status: number;
}

export interface CredentialRequest {
    publicKey: PublicKeyCredentialRequestOptionsJSON;
}

/* eslint-disable no-unused-vars */
export enum AttestationResult {
    Success = 1,
    Failure,
    FailureExcluded,
    FailureUserConsent,
    FailureUserVerificationOrResidentKey,
    FailureSyntax,
    FailureSupport,
    FailureUnknown,
    FailureWebAuthnNotSupported,
    FailureToken,
}

export interface RegistrationResult {
    response?: RegistrationResponseJSON;
    result: AttestationResult;
}

export enum AssertionResult {
    Success = 1,
    Failure,
    FailureUserConsent,
    FailureU2FFacetID,
    FailureSyntax,
    FailureUnknown,
    FailureUnknownSecurity,
    FailureWebAuthnNotSupported,
    FailureChallenge,
    FailureUnrecognized,
}

export function AssertionResultFailureString(result: AssertionResult) {
    switch (result) {
        case AssertionResult.Success:
            return "";
        case AssertionResult.FailureUserConsent:
            return "You cancelled the assertion request";
        case AssertionResult.FailureU2FFacetID:
            return "The server responded with an invalid Facet ID for the URL";
        case AssertionResult.FailureSyntax:
            return "The assertion challenge was rejected as malformed or incompatible by your browser";
        case AssertionResult.FailureWebAuthnNotSupported:
            return "Your browser does not support the WebAuthn protocol";
        case AssertionResult.FailureUnrecognized:
            return "This device is not registered";
        case AssertionResult.FailureUnknownSecurity:
            return "An unknown security error occurred";
        case AssertionResult.FailureUnknown:
            return "An unknown error occurred";
        default:
            return "An unexpected error occurred";
    }
}

export function AttestationResultFailureString(result: AttestationResult) {
    switch (result) {
        case AttestationResult.FailureToken:
            return "You must open the link from the same device and browser that initiated the registration process";
        case AttestationResult.FailureSupport:
            return "Your browser does not appear to support the configuration";
        case AttestationResult.FailureSyntax:
            return "The attestation challenge was rejected as malformed or incompatible by your browser";
        case AttestationResult.FailureWebAuthnNotSupported:
            return "Your browser does not support the WebAuthn protocol";
        case AttestationResult.FailureUserConsent:
            return "You cancelled the attestation request";
        case AttestationResult.FailureUserVerificationOrResidentKey:
            return "Your device does not support user verification or resident keys but this was required";
        case AttestationResult.FailureExcluded:
            return "You have registered this device already";
        case AttestationResult.FailureUnknown:
            return "An unknown error occurred";
    }

    return "";
}

export interface AuthenticationResult {
    response?: AuthenticationResponseJSON;
    result: AssertionResult;
}

export interface WebAuthnCredential {
    id: string;
    created_at: string;
    last_used_at?: string;
    rpid: string;
    description: string;
    kid: Uint8Array;
    aaguid?: string;
    attestation_type: string;
    attachment: string;
    transports: null | string[];
    sign_count: number;
    clone_warning: boolean;
    legacy: boolean;
    discoverable: boolean;
    present: boolean;
    verified: boolean;
    backup_eligible: boolean;
    backup_state: boolean;
    public_key: Uint8Array;
}

export function toAttachmentName(attachment: string) {
    switch (attachment.toLowerCase()) {
        case "cross-platform":
            return "Cross-Platform";
        case "platform":
            return "Platform";
        default:
            return "Unknown";
    }
}

export function toTransportName(transport: string) {
    switch (transport.toLowerCase()) {
        case "internal":
            return "Internal";
        case "ble":
            return "Bluetooth";
        case "nfc":
        case "usb":
            return transport.toUpperCase();
        default:
            return transport;
    }
}

export enum WebAuthnTouchState {
    WaitTouch = 1,
    InProgress = 2,
    Failure = 3,
}
