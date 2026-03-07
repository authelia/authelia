import { TOTPAlgorithm, toAlgorithmString, toEnum } from "@models/TOTPConfiguration";

it("converts algorithm enum to string", () => {
    expect(toAlgorithmString(TOTPAlgorithm.SHA1)).toBe("SHA1");
    expect(toAlgorithmString(TOTPAlgorithm.SHA256)).toBe("SHA256");
    expect(toAlgorithmString(TOTPAlgorithm.SHA512)).toBe("SHA512");
});

it("converts algorithm string to enum", () => {
    expect(toEnum("SHA1")).toBe(TOTPAlgorithm.SHA1);
    expect(toEnum("SHA256")).toBe(TOTPAlgorithm.SHA256);
    expect(toEnum("SHA512")).toBe(TOTPAlgorithm.SHA512);
});
