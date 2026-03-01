import {
    AuthenticationResponseJSON,
    PublicKeyCredentialCreationOptionsJSON,
    PublicKeyCredentialRequestOptionsJSON,
    RegistrationResponseJSON,
    startAuthentication,
    startRegistration,
} from "@simplewebauthn/browser";
import axios, { AxiosError, AxiosResponse } from "axios";

import {
    AssertionResult,
    AttestationResult,
    AuthenticationResult,
    CredentialCreation,
    CredentialRequest,
    PublicKeyCredentialCreationOptionsStatus,
    PublicKeyCredentialRequestOptionsStatus,
    RegistrationResult,
} from "@models/WebAuthn";
import {
    AuthenticationOKResponse,
    FirstFactorPasskeyPath,
    OptionalDataServiceResponse,
    ServiceResponse,
    WebAuthnAssertionPath,
    WebAuthnCredentialPath,
    WebAuthnRegistrationPath,
    validateStatusAuthentication,
    validateStatusWebAuthnCreation,
} from "@services/Api";
import { SignInResponse } from "@services/SignIn";

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
        case "AbortError":
        case "NotAllowedError":
            // § 6.3.2 Step 3 and Step 6.
            return AttestationResult.FailureUserConsent;
        case "ConstraintError":
            // § 6.3.2 Step 4.
            return AttestationResult.FailureUserVerificationOrResidentKey;
        default:
            console.error(`Unhandled DOMException occurred during WebAuthn attestation: ${exception}`);
            return AttestationResult.FailureUnknown;
    }
}

function getAssertionResultFromDOMException(
    exception: DOMException,
    options: PublicKeyCredentialRequestOptionsJSON,
): AssertionResult {
    // Docs for this section:
    // https://w3c.github.io/webauthn/#sctn-op-get-assertion
    switch (exception.name) {
        case "UnknownError":
            // § 6.3.3 Step 1 and Step 12.
            return AssertionResult.FailureSyntax;
        case "InvalidStateError":
            // § 6.3.2 Step 3.
            return AssertionResult.FailureUnrecognized;
        case "AbortError":
        case "NotAllowedError":
            // § 6.3.3 Step 6 and Step 7.
            return AssertionResult.FailureUserConsent;
        case "SecurityError":
            if (options.extensions?.appid) {
                // § 10.1 and 10.2 Step 3.
                return AssertionResult.FailureU2FFacetID;
            } else {
                return AssertionResult.FailureUnknownSecurity;
            }
        default:
            console.error(`Unhandled DOMException occurred during WebAuthn assertion: ${exception}`);
            return AssertionResult.FailureUnknown;
    }
}

export async function getWebAuthnOptions(): Promise<PublicKeyCredentialRequestOptionsStatus> {
    let response: AxiosResponse<ServiceResponse<CredentialRequest>>;

    response = await axios.get<ServiceResponse<CredentialRequest>>(WebAuthnAssertionPath);

    if (response.data.status !== "OK" || response.data.data == null) {
        return {
            status: response.status,
        };
    }

    return {
        options: response.data.data.publicKey,
        status: response.status,
    };
}

export async function getWebAuthnResult(options: PublicKeyCredentialRequestOptionsJSON) {
    const result: AuthenticationResult = {
        result: AssertionResult.Success,
    };

    try {
        result.response = await startAuthentication({ optionsJSON: options, useBrowserAutofill: true });
    } catch (e) {
        const exception = e as DOMException;
        if (exception) {
            result.result = getAssertionResultFromDOMException(exception, options);

            console.error(exception);

            return result;
        } else {
            console.error(`Unhandled exception occurred during WebAuthn authentication: ${e}`);
        }
    }

    if (result.response == null) {
        result.result = AssertionResult.Failure;
    } else {
        result.result = AssertionResult.Success;
    }

    return result;
}

export async function postWebAuthnResponse(
    response: AuthenticationResponseJSON,
    targetURL?: string | undefined,
    flowID?: string,
    flow?: string,
    subflow?: string,
    userCode?: string,
) {
    return axios.post<ServiceResponse<SignInResponse>>(WebAuthnAssertionPath, {
        flow,
        flowID,
        response,
        subflow,
        targetURL,
        userCode,
    });
}

export async function getWebAuthnPasskeyOptions(): Promise<PublicKeyCredentialRequestOptionsStatus> {
    let response: AxiosResponse<ServiceResponse<CredentialRequest>>;

    response = await axios.get<ServiceResponse<CredentialRequest>>(FirstFactorPasskeyPath);

    if (response.data.status !== "OK" || response.data.data == null) {
        return {
            status: response.status,
        };
    }

    return {
        options: response.data.data.publicKey,
        status: response.status,
    };
}

interface PostFirstFactorPasskeyBody {
    response: AuthenticationResponseJSON;
    keepMeLoggedIn: boolean;
    targetURL?: string;
    requestMethod?: string;
    flowID?: string;
    flow?: string;
    subflow?: string;
}

export async function postWebAuthnPasskeyResponse(
    response: AuthenticationResponseJSON,
    keepMeLoggedIn: boolean,
    targetURL?: string | undefined,
    requestMethod?: string,
    flowID?: string,
    flow?: string,
    subflow?: string,
) {
    const data: PostFirstFactorPasskeyBody = {
        flow,
        flowID,
        keepMeLoggedIn,
        requestMethod,
        response,
        subflow,
        targetURL,
    };

    return axios.post<ServiceResponse<SignInResponse>>(FirstFactorPasskeyPath, data);
}

export async function getWebAuthnRegistrationOptions(
    description: string,
): Promise<PublicKeyCredentialCreationOptionsStatus> {
    const response = await axios.put<ServiceResponse<CredentialCreation>>(
        WebAuthnRegistrationPath,
        {
            description: description,
        },
        {
            validateStatus: validateStatusWebAuthnCreation,
        },
    );

    if (response.data.status !== "OK" || response.data.data == null) {
        return {
            status: response.status,
        };
    }

    return {
        options: response.data.data.publicKey,
        status: response.status,
    };
}

export async function startWebAuthnRegistration(options: PublicKeyCredentialCreationOptionsJSON) {
    const result: RegistrationResult = {
        result: AttestationResult.Failure,
    };

    try {
        result.response = await startRegistration({ optionsJSON: options });
    } catch (e) {
        const exception = e as DOMException;
        if (exception) {
            result.result = getAttestationResultFromDOMException(exception);
            console.error(exception);
            return result;
        } else {
            console.error(`Unhandled exception occurred during WebAuthn attestation: ${e}`);
        }
    }

    if (result.response != null) {
        result.result = AttestationResult.Success;
    }

    return result;
}

async function postWebAuthnRegistrationResponse(
    response: RegistrationResponseJSON,
): Promise<AxiosResponse<OptionalDataServiceResponse<any>>> {
    return axios.post<OptionalDataServiceResponse<any>>(WebAuthnRegistrationPath, response);
}

export async function finishWebAuthnRegistration(response: RegistrationResponseJSON) {
    let result = {
        message: "Device registration failed.",
        status: AttestationResult.Failure,
    };

    try {
        const resp = await postWebAuthnRegistrationResponse(response);
        if (resp.data.status === "OK" && (resp.status === 200 || resp.status === 201)) {
            return {
                message: "",
                status: AttestationResult.Success,
            };
        }
    } catch (error) {
        if (error instanceof AxiosError && error.response !== undefined) {
            result.message = error.response.data.message;
        }
    }

    return result;
}

export async function deleteUserWebAuthnCredential(credentialID: string) {
    return axios<AuthenticationOKResponse>({
        method: "DELETE",
        url: `${WebAuthnCredentialPath}/${credentialID}`,
        validateStatus: validateStatusAuthentication,
    });
}

export async function updateUserWebAuthnCredential(credentialID: string, description: string) {
    return axios<AuthenticationOKResponse>({
        data: { description: description },
        method: "PUT",
        url: `${WebAuthnCredentialPath}/${credentialID}`,
        validateStatus: validateStatusAuthentication,
    });
}
