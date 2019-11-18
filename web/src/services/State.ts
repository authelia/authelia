import { Get } from "./Client";
import { StatePath } from "./Api";

export enum AuthenticationLevel {
    Unauthenticated = 0,
    OneFactor = 1,
    TwoFactor = 2,
}

export interface AutheliaState {
    username: string;
    authentication_level: AuthenticationLevel
}

export function getState() {
    return Get<AutheliaState>(StatePath);
}