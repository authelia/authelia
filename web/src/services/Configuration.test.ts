// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { SecondFactorMethod } from "@models/Methods";
import { Get } from "@services/Client";
import { getConfiguration } from "@services/Configuration";
import { toSecondFactorMethod } from "@services/UserInfo";

vi.mock("@services/Api", () => ({
    ConfigurationPath: "/configuration",
}));
vi.mock("@services/Client", () => ({
    Get: vi.fn(),
}));
vi.mock("@services/UserInfo", () => ({
    toSecondFactorMethod: vi.fn((m: string) => m),
}));

it("gets configuration and transforms available methods", async () => {
    (Get as any).mockResolvedValue({
        available_methods: ["totp", "webauthn"],
        password_change_disabled: false,
        password_reset_disabled: true,
        recovery_codes_disabled: true,
    });
    (toSecondFactorMethod as any).mockImplementation((m: string) =>
        m === "totp" ? SecondFactorMethod.TOTP : SecondFactorMethod.WebAuthn,
    );

    const result = await getConfiguration();
    expect(Get).toHaveBeenCalledWith("/configuration");
    expect(result.available_methods).toBeInstanceOf(Set);
    expect(result.available_methods.has(SecondFactorMethod.TOTP)).toBe(true);
    expect(result.available_methods.has(SecondFactorMethod.WebAuthn)).toBe(true);
    expect(result.password_change_disabled).toBe(false);
    expect(result.password_reset_disabled).toBe(true);
});

it("defaults the app authenticator registration to false", async () => {
    (Get as any).mockResolvedValue({
        available_methods: [],
        password_change_disabled: false,
        password_reset_disabled: false,
    });

    const result = await getConfiguration();
    expect(result.app_authenticator_registration).toBe(false);
});

it("passes the app authenticator registration through", async () => {
    (Get as any).mockResolvedValue({
        app_authenticator_registration: true,
        available_methods: [],
        password_change_disabled: false,
        password_reset_disabled: false,
    });

    const result = await getConfiguration();
    expect(result.app_authenticator_registration).toBe(true);
});
