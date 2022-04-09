import axios from "axios";

import { ServiceResponse, hasServiceError, toData } from "@services/Api";

export async function PostWithOptionalResponse<T = undefined>(path: string, body?: any): Promise<T | undefined> {
    const res = await axios.post<ServiceResponse<T>>(path, body);

    if (res.status !== 200 || hasServiceError(res).errored) {
        // in order for i18n to be used, it is necessary to return the raw message received from the api
        throw new Error(`${hasServiceError(res).message}`);
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

export async function Get<T = undefined>(path: string): Promise<T> {
    const res = await axios.get<ServiceResponse<T>>(path);

    if (res.status !== 200 || hasServiceError(res).errored) {
        // in order for i18n to be used, it is necessary to return the raw message received from the api 
        throw new Error(`${hasServiceError(res).message}`);
    }

    const d = toData<T>(res);
    if (!d) {
        throw new Error("unexpected type of response");
    }
    return d;
}
