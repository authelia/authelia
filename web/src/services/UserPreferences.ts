import { Get, PostWithOptionalResponse } from "./Client";
import { UserPreferencesPath } from "./Api";
import { SecondFactorMethod } from "../models/Methods";
import { UserPreferences } from "../models/UserPreferences";

export type Method2FA = "u2f" | "totp" | "duo_push";

export interface UserPreferencesPayload {
    method: Method2FA;
}

export function toEnum(method: Method2FA): SecondFactorMethod {
    switch (method) {
        case "u2f":
            return SecondFactorMethod.U2F;
        case "totp":
            return SecondFactorMethod.TOTP;
        case "duo_push":
            return SecondFactorMethod.Duo;
    }
}

export function toString(method: SecondFactorMethod): Method2FA {
    switch (method) {
        case SecondFactorMethod.U2F:
            return "u2f";
        case SecondFactorMethod.TOTP:
            return "totp";
        case SecondFactorMethod.Duo:
            return "duo_push";
    }
}

export async function getUserPreferences(): Promise<UserPreferences> {
    const res = await Get<UserPreferencesPayload>(UserPreferencesPath);
    return { method: toEnum(res.method) };
}

export function setPrefered2FAMethod(method: SecondFactorMethod) {
    return PostWithOptionalResponse(UserPreferencesPath,
        { method: toString(method) } as UserPreferencesPayload);
}