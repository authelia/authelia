import {
    CompleteDuoDeviceSelectionPath,
    CompletePushNotificationSignInPath,
    InitiateDuoDeviceSelectionPath,
} from "@services/Api";
import { Get, PostWithOptionalResponse, PostWithOptionalResponseRateLimited } from "@services/Client";

interface CompletePushSignInBody {
    targetURL?: string;
    flowID?: string;
    flow?: string;
    subflow?: string;
    userCode?: string;
}

export function completePushNotificationSignIn(
    targetURL?: string,
    flowID?: string,
    flow?: string,
    subflow?: string,
    userCode?: string,
    signal?: AbortSignal,
) {
    const body: CompletePushSignInBody = {
        flow,
        flowID,
        subflow,
        targetURL,
        userCode,
    };

    return PostWithOptionalResponseRateLimited<DuoSignInResponse>(CompletePushNotificationSignInPath, body, signal);
}

export interface DuoSignInResponse {
    result: string;
    devices: DuoDevice[];
    redirect: string;
    enroll_url: string;
}

export interface DuoDevicesGetResponse {
    result: string;
    devices: DuoDevice[];
    enroll_url: string;
    preferred_device?: string;
    preferred_method?: string;
}

export interface DuoDevice {
    device: string;
    display_name: string;
    capabilities: string[];
}

export async function initiateDuoDeviceSelectionProcess(signal?: AbortSignal) {
    return Get<DuoDevicesGetResponse>(InitiateDuoDeviceSelectionPath, signal);
}

export async function getPreferredDuoDevice(signal?: AbortSignal) {
    return Get<DuoDevicesGetResponse>(CompletePushNotificationSignInPath, signal);
}

export interface DuoDevicePostRequest {
    device: string;
    method: string;
}

export async function completeDuoDeviceSelectionProcess(device: DuoDevicePostRequest, signal?: AbortSignal) {
    return PostWithOptionalResponse(
        CompleteDuoDeviceSelectionPath,
        { device: device.device, method: device.method },
        signal,
    );
}
