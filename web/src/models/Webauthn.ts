import {
    AuthenticationResponseJSON,
    PublicKeyCredentialCreationOptionsJSON,
    PublicKeyCredentialRequestOptionsJSON,
    RegistrationResponseJSON,
} from "@simplewebauthn/typescript-types";

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

export enum AttestationResult {
    Success = 1,
    Failure,
    FailureExcluded,
    FailureUserConsent,
    FailureUserVerificationOrResidentKey,
    FailureSyntax,
    FailureSupport,
    FailureUnknown,
    FailureWebauthnNotSupported,
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
    FailureWebauthnNotSupported,
    FailureChallenge,
    FailureUnrecognized,
}

export function AssertionResultFailureString(result: AssertionResult) {
    switch (result) {
        case AssertionResult.Success:
            return "";
        case AssertionResult.FailureUserConsent:
            return "You cancelled the assertion request.";
        case AssertionResult.FailureU2FFacetID:
            return "The server responded with an invalid Facet ID for the URL.";
        case AssertionResult.FailureSyntax:
            return "The assertion challenge was rejected as malformed or incompatible by your browser.";
        case AssertionResult.FailureWebauthnNotSupported:
            return "Your browser does not support the WebAuthN protocol.";
        case AssertionResult.FailureUnrecognized:
            return "This device is not registered.";
        case AssertionResult.FailureUnknownSecurity:
            return "An unknown security error occurred.";
        case AssertionResult.FailureUnknown:
            return "An unknown error occurred.";
        default:
            return "An unexpected error occurred.";
    }
}

export function AttestationResultFailureString(result: AttestationResult) {
    switch (result) {
        case AttestationResult.FailureToken:
            return "You must open the link from the same device and browser that initiated the registration process.";
        case AttestationResult.FailureSupport:
            return "Your browser does not appear to support the configuration.";
        case AttestationResult.FailureSyntax:
            return "The attestation challenge was rejected as malformed or incompatible by your browser.";
        case AttestationResult.FailureWebauthnNotSupported:
            return "Your browser does not support the WebAuthN protocol.";
        case AttestationResult.FailureUserConsent:
            return "You cancelled the attestation request.";
        case AttestationResult.FailureUserVerificationOrResidentKey:
            return "Your device does not support user verification or resident keys but this was required.";
        case AttestationResult.FailureExcluded:
            return "You have registered this device already.";
        case AttestationResult.FailureUnknown:
            return "An unknown error occurred.";
    }

    return "";
}

export interface AuthenticationResult {
    response?: AuthenticationResponseJSON;
    result: AssertionResult;
}

export interface WebauthnDevice {
    id: string;
    created_at: Date;
    last_used_at?: Date;
    rpid: string;
    description: string;
    kid: Uint8Array;
    public_key: Uint8Array;
    attestation_type: string;
    transports: string[];
    aaguid?: string;
    sign_count: number;
    clone_warning: boolean;
}

export enum WebauthnTouchState {
    WaitTouch = 1,
    InProgress = 2,
    Failure = 3,
}
