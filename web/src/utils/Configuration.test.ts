import {
    getDuoSelfEnrollment,
    getEmbeddedVariable,
    getLogoOverride,
    getPasskeyLogin,
    getPrivacyPolicyEnabled,
    getPrivacyPolicyRequireAccept,
    getPrivacyPolicyURL,
    getRememberMe,
    getResetPassword,
    getResetPasswordCustomURL,
    getTheme,
} from "@utils/Configuration";

beforeEach(() => {
    document.body.getAttributeNames().forEach((attr) => document.body.removeAttribute(attr));
});

it("returns the embedded variable value", () => {
    document.body.setAttribute("data-testvar", "hello");
    expect(getEmbeddedVariable("testvar")).toBe("hello");
});

it("throws when the embedded variable is missing", () => {
    expect(() => getEmbeddedVariable("missing")).toThrow("No missing embedded variable detected");
});

it("returns true when duo self enrollment is enabled", () => {
    document.body.setAttribute("data-duoselfenrollment", "true");
    expect(getDuoSelfEnrollment()).toBe(true);
});

it("returns false when duo self enrollment is not true", () => {
    document.body.setAttribute("data-duoselfenrollment", "false");
    expect(getDuoSelfEnrollment()).toBe(false);
});

it("returns true when logo override is enabled", () => {
    document.body.setAttribute("data-logooverride", "true");
    expect(getLogoOverride()).toBe(true);
});

it("returns false when logo override is not true", () => {
    document.body.setAttribute("data-logooverride", "false");
    expect(getLogoOverride()).toBe(false);
});

it("returns true when remember me is enabled", () => {
    document.body.setAttribute("data-rememberme", "true");
    expect(getRememberMe()).toBe(true);
});

it("returns true when reset password is enabled", () => {
    document.body.setAttribute("data-resetpassword", "true");
    expect(getResetPassword()).toBe(true);
});

it("returns true when passkey login is enabled", () => {
    document.body.setAttribute("data-passkeylogin", "true");
    expect(getPasskeyLogin()).toBe(true);
});

it("returns the reset password custom URL", () => {
    document.body.setAttribute("data-resetpasswordcustomurl", "https://example.com");
    expect(getResetPasswordCustomURL()).toBe("https://example.com");
});

it("returns true when privacy policy URL is not empty", () => {
    document.body.setAttribute("data-privacypolicyurl", "https://example.com/privacy");
    expect(getPrivacyPolicyEnabled()).toBe(true);
});

it("returns false when privacy policy URL is empty", () => {
    document.body.setAttribute("data-privacypolicyurl", "");
    expect(getPrivacyPolicyEnabled()).toBe(false);
});

it("returns the privacy policy URL", () => {
    document.body.setAttribute("data-privacypolicyurl", "https://example.com/privacy");
    expect(getPrivacyPolicyURL()).toBe("https://example.com/privacy");
});

it("returns true when privacy policy accept is required", () => {
    document.body.setAttribute("data-privacypolicyaccept", "true");
    expect(getPrivacyPolicyRequireAccept()).toBe(true);
});

it("returns the theme value", () => {
    document.body.setAttribute("data-theme", "dark");
    expect(getTheme()).toBe("dark");
});
