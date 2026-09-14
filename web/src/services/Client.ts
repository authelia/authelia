// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import axios from "axios";

import {
    RateLimitedData,
    ServiceResponse,
    hasServiceError,
    toData,
    toDataRateLimited,
    validateStatusTooManyRequests,
} from "@services/Api";
import { ServiceError } from "@services/ServiceError";

export async function PutWithOptionalResponse<T = undefined>(
    path: string,
    body?: any,
    signal?: AbortSignal,
): Promise<T | undefined> {
    const res = await axios.put<ServiceResponse<T>>(path, body, { signal });

    if (res.status !== 200 || hasServiceError(res).errored) {
        const { code, message } = hasServiceError(res);

        throw new ServiceError(`Failed PUT to ${path}. Code: ${res.status}. Message: ${message}`, res.status, code);
    }

    return toData<T>(res);
}

export async function PostWithOptionalResponse<T = undefined>(
    path: string,
    body?: any,
    signal?: AbortSignal,
): Promise<T | undefined> {
    const res = await axios.post<ServiceResponse<T>>(path, body, { signal });

    if (res.status !== 200 || hasServiceError(res).errored) {
        const { code, message } = hasServiceError(res);

        throw new ServiceError(`Failed POST to ${path}. Code: ${res.status}. Message: ${message}`, res.status, code);
    }

    return toData<T>(res);
}

export async function PostWithOptionalResponseRateLimited<T = undefined>(
    path: string,
    body?: any,
    signal?: AbortSignal,
): Promise<RateLimitedData<T> | undefined> {
    const res = await axios.post<ServiceResponse<T>>(path, body, {
        signal,
        validateStatus: validateStatusTooManyRequests,
    });

    if (res.status !== 200 || hasServiceError(res).errored) {
        if (res.status === 429) {
            return toDataRateLimited<T>(res);
        }

        const { code, message } = hasServiceError(res);

        throw new ServiceError(`Failed POST to ${path}. Code: ${res.status}. Message: ${message}`, res.status, code);
    }

    return toDataRateLimited<T>(res);
}

export async function DeleteWithOptionalResponse<T = undefined>(
    path: string,
    body?: any,
    signal?: AbortSignal,
): Promise<T | undefined> {
    const res = await axios.delete<ServiceResponse<T>>(path, { data: body, signal });

    if (res.status !== 200 || hasServiceError(res).errored) {
        const { code, message } = hasServiceError(res);

        throw new ServiceError(`Failed DELETE to ${path}. Code: ${res.status}. Message: ${message}`, res.status, code);
    }

    return toData<T>(res);
}

export async function Post<T>(path: string, body?: any, signal?: AbortSignal) {
    const res = await PostWithOptionalResponse<T>(path, body, signal);

    if (!res) {
        throw new Error("unexpected type of response");
    }

    return res;
}

export async function Put<T>(path: string, body?: any, signal?: AbortSignal) {
    const res = await PutWithOptionalResponse<T>(path, body, signal);

    if (!res) {
        throw new Error("unexpected type of response");
    }

    return res;
}

export async function Get<T = undefined>(path: string, signal?: AbortSignal): Promise<T> {
    const res = await axios.get<ServiceResponse<T>>(path, { signal });

    if (res.status !== 200 || hasServiceError(res).errored) {
        throw new ServiceError(`Failed GET from ${path}. Code: ${res.status}.`, res.status);
    }

    const d = toData<T>(res);

    if (!d) {
        throw new Error("unexpected type of response");
    }

    return d;
}

export async function GetWithOptionalData<T = undefined>(path: string, signal?: AbortSignal): Promise<null | T> {
    const res = await axios.get<ServiceResponse<T>>(path, { signal });

    if (res.status !== 200 || hasServiceError(res).errored) {
        throw new ServiceError(`Failed GET from ${path}. Code: ${res.status}.`, res.status);
    }

    const d = toData<T>(res);

    if (d === null) {
        return null;
    }

    if (!d) {
        throw new Error("unexpected type of response");
    }

    return d;
}
