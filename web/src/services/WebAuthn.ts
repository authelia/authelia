import { startAuthentication, startRegistration } from "@simplewebauthn/browser";
import {
    AuthenticationResponseJSON,
    PublicKeyCredentialCreationOptionsJSON,
    PublicKeyCredentialRequestOptionsJSON,
    RegistrationResponseJSON,
} from "@simplewebauthn/types";
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
            if (options.extensions?.appid !== undefined) {
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
        result.response = await startAuthentication(options);
    } catch (e) {
        const exception = e as DOMException;
        if (exception !== undefined) {
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
    workflow?: string,
    workflowID?: string,
) {
    return axios.post<ServiceResponse<SignInResponse>>(WebAuthnAssertionPath, {
        response: response,
        targetURL: targetURL,
        workflow: workflow,
        workflowID: workflowID,
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
    workflow?: string;
}

export async function postWebAuthnPasskeyResponse(
    response: AuthenticationResponseJSON,
    keepMeLoggedIn: boolean,
    targetURL?: string | undefined,
    requestMethod?: string,
    workflow?: string,
) {
    const data: PostFirstFactorPasskeyBody = {
        response,
        keepMeLoggedIn,
    };

    if (data.response.response.userHandle) {
        // Encode the userHandle to match the typing on the backend.
        data.response.response.userHandle = btoa(data.response.response.userHandle)
            .replace(/\+/g, "-")
            .replace(/\//g, "_")
            .replace(/=/g, "");
    }

    if (targetURL) {
        data.targetURL = targetURL;
    }

    if (requestMethod) {
        data.requestMethod = requestMethod;
    }

    if (workflow) {
        data.workflow = workflow;
    }

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
        result.response = await startRegistration(options);
    } catch (e) {
        const exception = e as DOMException;
        if (exception !== undefined) {
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
        status: AttestationResult.Failure,
        message: "Device registration failed.",
    };

    try {
        const resp = await postWebAuthnRegistrationResponse(response);
        if (resp.data.status === "OK" && (resp.status === 200 || resp.status === 201)) {
            return {
                status: AttestationResult.Success,
                message: "",
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
        method: "PUT",
        url: `${WebAuthnCredentialPath}/${credentialID}`,
        data: { description: description },
        validateStatus: validateStatusAuthentication,
    });
}
