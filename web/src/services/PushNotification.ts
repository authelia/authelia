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
}

export function completePushNotificationSignIn(targetURL?: string, flowID?: string, flow?: string, subflow?: string) {
    const body: CompletePushSignInBody = {
        targetURL,
        flowID,
        flow,
        subflow,
    };

    return PostWithOptionalResponseRateLimited<DuoSignInResponse>(CompletePushNotificationSignInPath, body);
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
}

export interface DuoDevice {
    device: string;
    display_name: string;
    capabilities: string[];
}

export async function initiateDuoDeviceSelectionProcess() {
    return Get<DuoDevicesGetResponse>(InitiateDuoDeviceSelectionPath);
}

export interface DuoDevicePostRequest {
    device: string;
    method: string;
}

export async function completeDuoDeviceSelectionProcess(device: DuoDevicePostRequest) {
    return PostWithOptionalResponse(CompleteDuoDeviceSelectionPath, { device: device.device, method: device.method });
}
