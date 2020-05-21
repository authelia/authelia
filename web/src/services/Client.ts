import axios from "axios";
import { ServiceResponse, hasServiceError, toData } from "./Api";

export async function PostWithOptionalResponse<T = undefined>(path: string, body?: any) {
    const res = await axios.post<ServiceResponse<T>>(path, body);

    if (res.status !== 200 || hasServiceError(res)) {
        throw new Error(`Failed POST to ${path}. Code: ${res.status}.`);
    }
    return toData(res);
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

    if (res.status !== 200 || hasServiceError(res)) {
        throw new Error(`Failed GET from ${path}. Code: ${res.status}.`);
    }

    const d = toData<T>(res);
    if (!d) {
        throw new Error("unexpected type of response");
    }
    return d;
}