import { GetWithOptionalData } from "@services/Client";
import { getUserWebAuthnCredentials } from "@services/UserWebAuthnCredentials";

vi.mock("@services/Api", () => ({
    WebAuthnCredentialsPath: "/webauthn/credentials",
}));
vi.mock("@services/Client", () => ({
    GetWithOptionalData: vi.fn(),
}));

it("returns credentials when present", async () => {
    const creds = [{ description: "key", id: "1" }];
    (GetWithOptionalData as any).mockResolvedValue(creds);
    const result = await getUserWebAuthnCredentials();
    expect(GetWithOptionalData).toHaveBeenCalledWith("/webauthn/credentials");
    expect(result).toEqual(creds);
});

it("returns empty array when null", async () => {
    (GetWithOptionalData as any).mockResolvedValue(null);
    const result = await getUserWebAuthnCredentials();
    expect(result).toEqual([]);
});
