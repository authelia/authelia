import { StatePath } from "@services/Api";
import { Get } from "@services/Client";

/* eslint-disable no-unused-vars */
export enum AuthenticationLevel {
    Unauthenticated = 0,
    OneFactor = 1,
    TwoFactor = 2,
}

export interface AutheliaState {
    username: string;
    authentication_level: AuthenticationLevel;
    factor_knowledge: boolean;
    default_redirection_url?: string;
}

export async function getState(): Promise<AutheliaState> {
    return Get<AutheliaState>(StatePath);
}
