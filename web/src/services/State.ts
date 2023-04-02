// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { StatePath } from "@services/Api";
import { Get } from "@services/Client";

export enum AuthenticationLevel {
    Unauthenticated = 0,
    OneFactor = 1,
    TwoFactor = 2,
}

export interface AutheliaState {
    username: string;
    authentication_level: AuthenticationLevel;
}

export async function getState(): Promise<AutheliaState> {
    return Get<AutheliaState>(StatePath);
}
