import {
    AssertionResult,
    AssertionResultFailureString,
    AttestationResult,
    AttestationResultFailureString,
    toAttachmentName,
    toTransportName,
} from "@models/WebAuthn";

it("returns correct strings for assertion result failures", () => {
    expect(AssertionResultFailureString(AssertionResult.Success)).toBe("");
    expect(AssertionResultFailureString(AssertionResult.FailureUserConsent)).toBe(
        "You cancelled the assertion request",
    );
    expect(AssertionResultFailureString(AssertionResult.FailureU2FFacetID)).toBe(
        "The server responded with an invalid Facet ID for the URL",
    );
    expect(AssertionResultFailureString(AssertionResult.FailureSyntax)).toBe(
        "The assertion challenge was rejected as malformed or incompatible by your browser",
    );
    expect(AssertionResultFailureString(AssertionResult.FailureWebAuthnNotSupported)).toBe(
        "Your browser does not support the WebAuthn protocol",
    );
    expect(AssertionResultFailureString(AssertionResult.FailureUnrecognized)).toBe("This device is not registered");
    expect(AssertionResultFailureString(AssertionResult.FailureUnknownSecurity)).toBe(
        "An unknown security error occurred",
    );
    expect(AssertionResultFailureString(AssertionResult.FailureUnknown)).toBe("An unknown error occurred");
    expect(AssertionResultFailureString(AssertionResult.Failure as any)).toBe("An unexpected error occurred");
});

it("returns correct strings for attestation result failures", () => {
    expect(AttestationResultFailureString(AttestationResult.FailureToken)).toBe(
        "You must open the link from the same device and browser that initiated the registration process",
    );
    expect(AttestationResultFailureString(AttestationResult.FailureSupport)).toBe(
        "Your browser does not appear to support the configuration",
    );
    expect(AttestationResultFailureString(AttestationResult.FailureSyntax)).toBe(
        "The attestation challenge was rejected as malformed or incompatible by your browser",
    );
    expect(AttestationResultFailureString(AttestationResult.FailureWebAuthnNotSupported)).toBe(
        "Your browser does not support the WebAuthn protocol",
    );
    expect(AttestationResultFailureString(AttestationResult.FailureUserConsent)).toBe(
        "You cancelled the attestation request",
    );
    expect(AttestationResultFailureString(AttestationResult.FailureUserVerificationOrResidentKey)).toBe(
        "Your device does not support user verification or resident keys but this was required",
    );
    expect(AttestationResultFailureString(AttestationResult.FailureExcluded)).toBe(
        "You have registered this device already",
    );
    expect(AttestationResultFailureString(AttestationResult.FailureUnknown)).toBe("An unknown error occurred");
    expect(AttestationResultFailureString(AttestationResult.Success)).toBe("");
});

it("returns correct names for attachments", () => {
    expect(toAttachmentName("cross-platform")).toBe("Cross-Platform");
    expect(toAttachmentName("platform")).toBe("Platform");
    expect(toAttachmentName("unknown")).toBe("Unknown");
    expect(toAttachmentName("CROSS-PLATFORM")).toBe("Cross-Platform");
});

it("returns correct names for transports", () => {
    expect(toTransportName("internal")).toBe("Internal");
    expect(toTransportName("ble")).toBe("Bluetooth");
    expect(toTransportName("nfc")).toBe("NFC");
    expect(toTransportName("usb")).toBe("USB");
    expect(toTransportName("unknown")).toBe("unknown");
    expect(toTransportName("INTERNAL")).toBe("Internal");
});
