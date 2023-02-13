import axios from "axios";

import { ServiceResponse, hasServiceError, toData } from "@services/Api";

export async function PutWithOptionalResponse<T = undefined>(path: string, body?: any): Promise<T | undefined> {
    const res = await axios.put<ServiceResponse<T>>(path, body);

    if (res.status !== 200 || hasServiceError(res).errored) {
        throw new Error(`Failed POST to ${path}. Code: ${res.status}. Message: ${hasServiceError(res).message}`);
    }

    return toData<T>(res);
}

export async function PostWithOptionalResponse<T = undefined>(path: string, body?: any): Promise<T | undefined> {
    const res = await axios.post<ServiceResponse<T>>(path, body);

    if (res.status !== 200 || hasServiceError(res).errored) {
        throw new Error(`Failed POST to ${path}. Code: ${res.status}. Message: ${hasServiceError(res).message}`);
    }

    return toData<T>(res);
}

export async function DeleteWithOptionalResponse<T = undefined>(path: string, body?: any): Promise<T | undefined> {
    const res = await axios.delete<ServiceResponse<T>>(path, body);

    if (res.status !== 200 || hasServiceError(res).errored) {
        throw new Error(`Failed DELETE to ${path}. Code: ${res.status}. Message: ${hasServiceError(res).message}`);
    }

    return toData<T>(res);
}

export async function Post<T>(path: string, body?: any) {
    const res = await PostWithOptionalResponse<T>(path, body);
    if (!res) {
        throw new Error("unexpected type of response");
    }
    return res;
}

export async function Put<T>(path: string, body?: any) {
    const res = await PutWithOptionalResponse<T>(path, body);
    if (!res) {
        throw new Error("unexpected type of response");
    }
    return res;
}

export async function Get<T = undefined>(path: string): Promise<T> {
    const res = await axios.get<ServiceResponse<T>>(path);

    if (res.status !== 200 || hasServiceError(res).errored) {
        throw new Error(`Failed GET from ${path}. Code: ${res.status}.`);
    }

    const d = toData<T>(res);
    if (!d) {
        throw new Error("unexpected type of response");
    }
    return d;
}

export async function GetWithOptionalData<T = undefined>(path: string): Promise<T | null> {
    const res = await axios.get<ServiceResponse<T>>(path);

    if (res.status !== 200 || hasServiceError(res).errored) {
        throw new Error(`Failed GET from ${path}. Code: ${res.status}.`);
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
