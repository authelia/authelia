import { vi } from "vitest";

import { PasswordPolicyMode } from "@models/PasswordPolicy";
import { Get } from "@services/Client";
import { getPasswordPolicyConfiguration, toEnum } from "@services/PasswordPolicyConfiguration";

vi.mock("@services/Api", () => ({
    PasswordPolicyConfigurationPath: "/password-policy",
}));
vi.mock("@services/Client", () => ({
    Get: vi.fn(),
}));

it("converts disabled to enum", () => {
    expect(toEnum("disabled")).toBe(PasswordPolicyMode.Disabled);
});

it("converts standard to enum", () => {
    expect(toEnum("standard")).toBe(PasswordPolicyMode.Standard);
});

it("converts zxcvbn to enum", () => {
    expect(toEnum("zxcvbn")).toBe(PasswordPolicyMode.ZXCVBN);
});

it("gets password policy configuration", async () => {
    const mockConfig = {
        mode: "standard" as const,
        min_length: 8,
        max_length: 128,
        min_score: 3,
        require_uppercase: true,
        require_lowercase: true,
        require_number: true,
        require_special: false,
    };
    (Get as any).mockResolvedValue(mockConfig);
    const result = await getPasswordPolicyConfiguration();
    expect(Get).toHaveBeenCalledWith("/password-policy");
    expect(result).toEqual({
        ...mockConfig,
        mode: PasswordPolicyMode.Standard,
    });
});
