import { Get } from "./Client";
import { Available2FAMethodsPath } from "./Api";
import { Method2FA, toEnum } from "./UserPreferences";
import { Configuration } from "../models/Configuration";

export async function getAvailable2FAMethods(): Promise<Configuration> {
    const methods = await Get<Method2FA[]>(Available2FAMethodsPath);
    return new Set(methods.map(toEnum));
}