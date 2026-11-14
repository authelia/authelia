// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import axios from "axios";

import {
    AppAuthenticatorDevicesPath,
    AppAuthenticatorRegistrationPath,
    AppAuthenticatorSignInPath,
    AppAuthenticatorSignInStatusPath,
    AuthenticationOKResponse,
    validateStatusAuthentication,
} from "@services/Api";
import {
    DeleteWithOptionalResponse,
    Get,
    GetWithOptionalData,
    Post,
    PostWithOptionalResponse,
    Put,
} from "@services/Client";

const officialAppIdentity = "com.authelia.mobile";

export interface AppAuthenticatorSignInBody {
    targetURL?: string;
    flowID?: string;
    flow?: string;
    subflow?: string;
    userCode?: string;
}

export interface AppAuthenticatorChallenge {
    number: string;
    nonce_length: number;
    expires_at: number;
}

export type AppAuthenticatorChallengeState = "denied" | "expired" | "pending" | "rejected";

export interface AppAuthenticatorSignInStatus {
    state?: AppAuthenticatorChallengeState;
    redirect?: string;
}

export interface AppAuthenticatorRegistrationStart {
    qr: string;
    expires_at: number;
}

export interface AppAuthenticatorRegistrationDevice {
    platform: string;
    model: string;
    name: string;
}

export type AppAuthenticatorRegistrationStatus =
    | {
          state: "confirm_number";
          number: string;
          nonce_length: number;
          device: AppAuthenticatorRegistrationDevice;
          expires_at: number;
      }
    | { state: "expired" }
    | { state: "failed" }
    | { state: "registered"; id?: number }
    | { state: "waiting"; expires_at: number };

export interface AppAuthenticatorDevice {
    id: number;
    created_at: string;
    last_used_at?: string;
    name: string;
    model: string;
    platform: string;
    attestation_format: string;
    attestation_identity?: string;
    attestation_detail?: string;
}

export type AppAuthenticatorAttestationStatus =
    { kind: "none" } | { kind: "official" } | { kind: "verified"; identity: string };

export async function initiateAppAuthenticatorSignIn(body: AppAuthenticatorSignInBody, signal?: AbortSignal) {
    return Post<AppAuthenticatorChallenge>(AppAuthenticatorSignInPath, body, signal);
}

export async function getAppAuthenticatorSignInStatus(body: AppAuthenticatorSignInBody, signal?: AbortSignal) {
    return PostWithOptionalResponse<AppAuthenticatorSignInStatus>(AppAuthenticatorSignInStatusPath, body, signal);
}

export async function startAppAuthenticatorRegistration(signal?: AbortSignal) {
    return Put<AppAuthenticatorRegistrationStart>(AppAuthenticatorRegistrationPath, undefined, signal);
}

export async function getAppAuthenticatorRegistration(signal?: AbortSignal) {
    return Get<AppAuthenticatorRegistrationStatus>(AppAuthenticatorRegistrationPath, signal);
}

export async function cancelAppAuthenticatorRegistration(signal?: AbortSignal) {
    return DeleteWithOptionalResponse(AppAuthenticatorRegistrationPath, undefined, signal);
}

export async function getAppAuthenticatorDevices(): Promise<AppAuthenticatorDevice[]> {
    const res = await GetWithOptionalData<AppAuthenticatorDevice[] | null>(AppAuthenticatorDevicesPath);

    return res ?? [];
}

export async function updateAppAuthenticatorDevice(id: number, name: string) {
    return axios<AuthenticationOKResponse>({
        data: { name },
        method: "PUT",
        url: `${AppAuthenticatorDevicesPath}/${id}`,
        validateStatus: validateStatusAuthentication,
    });
}

export async function deleteAppAuthenticatorDevice(id: number) {
    return axios<AuthenticationOKResponse>({
        method: "DELETE",
        url: `${AppAuthenticatorDevicesPath}/${id}`,
        validateStatus: validateStatusAuthentication,
    });
}

export function appAuthenticatorAttestationStatus(
    device: Pick<AppAuthenticatorDevice, "attestation_format" | "attestation_identity">,
): AppAuthenticatorAttestationStatus {
    const identity = device.attestation_identity;

    if (device.attestation_format === "none" || !identity) {
        return { kind: "none" };
    }

    const separator = identity.indexOf(".");
    const bundle = separator > 0 ? identity.slice(separator + 1) : "";

    if (
        (device.attestation_format === "android-key" && identity === officialAppIdentity) ||
        (device.attestation_format === "apple-appattest" && bundle === officialAppIdentity)
    ) {
        return { kind: "official" };
    }

    return { identity, kind: "verified" };
}
