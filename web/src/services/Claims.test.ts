import { setClaimCase } from "@services/Claims";

it("returns the claim unchanged", () => {
    expect(setClaimCase("test_claim")).toBe("test_claim");
    expect(setClaimCase("user_verified")).toBe("user_verified");
    expect(setClaimCase("first_second")).toBe("first_second");
    expect(setClaimCase("first second")).toBe("first second");
});
