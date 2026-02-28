export function getEmbeddedVariable(variableName: string) {
    const value = document.body.getAttribute(`data-${variableName}`);
    if (value === null) {
        throw new Error(`No ${variableName} embedded variable detected`);
    }

    return value;
}

export function getDuoSelfEnrollment() {
    return getEmbeddedVariable("duoselfenrollment") === "true";
}

export function getLogoOverride() {
    return getEmbeddedVariable("logooverride") === "true";
}

export function getRememberMe() {
    return getEmbeddedVariable("rememberme") === "true";
}

export function getResetPassword() {
    return getEmbeddedVariable("resetpassword") === "true";
}

export function getPasskeyLogin() {
    return getEmbeddedVariable("passkeylogin") === "true";
}

export function getSpnegoLogin() {
    return getEmbeddedVariable("spnegologin") === "true";
}

export function getResetPasswordCustomURL() {
    return getEmbeddedVariable("resetpasswordcustomurl");
}

export function getPrivacyPolicyEnabled() {
    return getEmbeddedVariable("privacypolicyurl") !== "";
}

export function getPrivacyPolicyURL() {
    return getEmbeddedVariable("privacypolicyurl");
}

export function getPrivacyPolicyRequireAccept() {
    return getEmbeddedVariable("privacypolicyaccept") === "true";
}

export function getTheme() {
    return getEmbeddedVariable("theme");
}
