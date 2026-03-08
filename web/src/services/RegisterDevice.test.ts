import { Put } from "@services/Client";
import { getTOTPSecret } from "@services/RegisterDevice";

vi.mock("@services/Api", () => ({
    TOTPRegistrationPath: "/totp/register",
}));
vi.mock("@services/Client", () => ({
    Put: vi.fn(),
}));

it("calls Put with algorithm, length, and period", async () => {
    (Put as any).mockResolvedValue({ base32_secret: "secret", otpauth_url: "url" });
    const result = await getTOTPSecret("SHA1", 6, 30);
    expect(Put).toHaveBeenCalledWith("/totp/register", { algorithm: "SHA1", length: 6, period: 30 });
    expect(result).toEqual({ base32_secret: "secret", otpauth_url: "url" });
});
