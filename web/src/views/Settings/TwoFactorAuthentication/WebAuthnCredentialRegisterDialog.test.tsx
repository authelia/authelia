// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import { AttestationResult } from "@models/WebAuthn";
import {
    finishWebAuthnRegistration,
    getWebAuthnRegistrationOptions,
    startWebAuthnRegistration,
} from "@services/WebAuthn";
import WebAuthnCredentialRegisterDialog from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialRegisterDialog";

const mocks = vi.hoisted(() => ({
    createErrorNotification: vi.fn(),
    createSuccessNotification: vi.fn(),
}));

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: mocks.createErrorNotification,
        createSuccessNotification: mocks.createSuccessNotification,
    }),
}));

vi.mock("@services/WebAuthn", () => ({
    finishWebAuthnRegistration: vi.fn(),
    getWebAuthnRegistrationOptions: vi.fn(),
    startWebAuthnRegistration: vi.fn(),
}));

vi.mock("@components/InformationIcon", () => ({
    default: () => <div data-testid="info-icon" />,
}));

vi.mock("@components/WebAuthnRegisterIcon", () => ({
    default: (props: any) => <div data-testid="webauthn-register-icon" data-timeout={props.timeout} />,
}));

const getOptionsMock = vi.mocked(getWebAuthnRegistrationOptions);
const startRegistrationMock = vi.mocked(startWebAuthnRegistration);
const finishRegistrationMock = vi.mocked(finishWebAuthnRegistration);

const creationOptions = { challenge: "challenge", timeout: 60000 } as any;

function renderDialog(props: Partial<{ open: boolean; setClosed: () => void }> = {}) {
    const setClosed = props.setClosed ?? vi.fn();
    return {
        ...render(<WebAuthnCredentialRegisterDialog open={props.open ?? true} setClosed={setClosed} />),
        setClosed,
    };
}

function getDescription() {
    return document.getElementById("webauthn-credential-description") as HTMLInputElement;
}

function clickNext() {
    fireEvent.click(document.getElementById("dialog-next") as HTMLButtonElement);
}

function clickCancel() {
    fireEvent.click(document.getElementById("dialog-cancel") as HTMLButtonElement);
}

beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, "error").mockImplementation(() => {});
    getOptionsMock.mockResolvedValue({ options: creationOptions, status: 200 } as any);
    startRegistrationMock.mockResolvedValue({
        response: { id: "credential-id" },
        result: AttestationResult.Success,
    } as any);
    finishRegistrationMock.mockResolvedValue({ status: AttestationResult.Success } as any);
});

afterEach(() => {
    vi.restoreAllMocks();
});

describe("rendering", () => {
    it("renders dialog with stepper when open", () => {
        renderDialog();
        expect(screen.getByText("Register {{item}}")).toBeInTheDocument();
        expect(screen.getAllByText("Description").length).toBeGreaterThanOrEqual(1);
        expect(screen.getByText("Verification")).toBeInTheDocument();
    });

    it("does not render content when closed", () => {
        renderDialog({ open: false });
        expect(screen.queryByText("Register {{item}}")).not.toBeInTheDocument();
    });

    it("shows the description prompt on the first step", () => {
        renderDialog();
        expect(screen.getByText("Enter a description for this WebAuthn Credential")).toBeInTheDocument();
        expect(screen.getByTestId("info-icon")).toBeInTheDocument();
        expect(getDescription()).toBeInTheDocument();
    });

    it("records the typed description", async () => {
        renderDialog();

        fireEvent.change(getDescription(), { target: { value: "My Key" } });

        await waitFor(() => expect(getDescription()).toHaveValue("My Key"));
    });
});

describe("description validation", () => {
    it("rejects an empty description", async () => {
        renderDialog();

        clickNext();

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith(
                "The Description must be more than 1 character and less than 64 characters",
            ),
        );
        expect(getDescription()).toHaveAttribute("aria-invalid", "true");
        expect(getOptionsMock).not.toHaveBeenCalled();
    });

    it("rejects a description that is only whitespace", async () => {
        renderDialog();

        fireEvent.change(getDescription(), { target: { value: "   " } });

        clickNext();

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith(
                "The Description must be more than 1 character and less than 64 characters",
            ),
        );
        expect(getDescription()).toHaveAttribute("aria-invalid", "true");
        expect(getOptionsMock).not.toHaveBeenCalled();
    });

    it("rejects a description longer than 64 characters", async () => {
        renderDialog();

        fireEvent.change(getDescription(), { target: { value: "a".repeat(65) } });
        clickNext();

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith(
                "The Description must be more than 1 character and less than 64 characters",
            ),
        );
        expect(getOptionsMock).not.toHaveBeenCalled();
    });

    it("clears the description error once the user edits the field", async () => {
        renderDialog();

        clickNext();
        await waitFor(() => expect(getDescription()).toHaveAttribute("aria-invalid", "true"));

        fireEvent.change(getDescription(), { target: { value: "My Key" } });

        await waitFor(() => expect(getDescription()).not.toHaveAttribute("aria-invalid", "true"));
    });

    it("does nothing when the dialog is closed", () => {
        renderDialog({ open: false });

        expect(getOptionsMock).not.toHaveBeenCalled();
    });
});

describe("obtaining the creation options", () => {
    it("requests options with the entered description", async () => {
        renderDialog();

        fireEvent.change(getDescription(), { target: { value: "My Key" } });
        clickNext();

        await waitFor(() => expect(getOptionsMock).toHaveBeenCalledWith("My Key"));
    });

    it("submits on Enter", async () => {
        renderDialog();

        fireEvent.change(getDescription(), { target: { value: "My Key" } });
        fireEvent.keyDown(getDescription(), { key: "Enter" });

        await waitFor(() => expect(getOptionsMock).toHaveBeenCalledWith("My Key"));
    });

    it("ignores keys other than Enter", () => {
        renderDialog();

        fireEvent.change(getDescription(), { target: { value: "My Key" } });
        fireEvent.keyDown(getDescription(), { key: "a" });

        expect(getOptionsMock).not.toHaveBeenCalled();
    });

    it("reports an empty options payload", async () => {
        getOptionsMock.mockResolvedValue({ options: null, status: 200 } as any);

        renderDialog();

        fireEvent.change(getDescription(), { target: { value: "My Key" } });
        clickNext();

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith(
                "Credential Creation Options Request succeeded but Credential Creation Options is empty",
            ),
        );
        expect(startRegistrationMock).not.toHaveBeenCalled();
    });

    it("reports a duplicate description", async () => {
        getOptionsMock.mockResolvedValue({ status: 409 } as any);

        renderDialog();

        fireEvent.change(getDescription(), { target: { value: "My Key" } });
        clickNext();

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith(
                "A WebAuthn Credential with that Description already exists",
            ),
        );
        expect(getDescription()).toHaveAttribute("aria-invalid", "true");
    });

    it("reports any other failure status", async () => {
        getOptionsMock.mockResolvedValue({ status: 500 } as any);

        renderDialog();

        fireEvent.change(getDescription(), { target: { value: "My Key" } });
        clickNext();

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith(
                "Error occurred obtaining the WebAuthn Credential creation options",
            ),
        );
    });
});

describe("credential creation", () => {
    it("registers the credential and reports success", async () => {
        const setClosed = vi.fn();

        renderDialog({ setClosed });

        fireEvent.change(getDescription(), { target: { value: "My Key" } });
        clickNext();

        await waitFor(() => expect(startRegistrationMock).toHaveBeenCalledWith(creationOptions));
        await waitFor(() => expect(finishRegistrationMock).toHaveBeenCalledWith({ id: "credential-id" }));
        await waitFor(() =>
            expect(mocks.createSuccessNotification).toHaveBeenCalledWith("Successfully {{action}} the {{item}}"),
        );
        expect(setClosed).toHaveBeenCalled();
    });

    it("reports a failing finish response", async () => {
        finishRegistrationMock.mockResolvedValue({
            message: "server rejected the credential",
            status: AttestationResult.Failure,
        } as any);

        renderDialog();

        fireEvent.change(getDescription(), { target: { value: "My Key" } });
        clickNext();

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith("server rejected the credential"),
        );
    });

    it("reports an empty registration response", async () => {
        startRegistrationMock.mockResolvedValue({ response: null, result: AttestationResult.Success } as any);

        renderDialog();

        fireEvent.change(getDescription(), { target: { value: "My Key" } });
        clickNext();

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith(
                "Credential Creation Request succeeded but Registration Response is empty.",
            ),
        );
        expect(finishRegistrationMock).not.toHaveBeenCalled();
    });

    it("reports a cancelled attestation", async () => {
        startRegistrationMock.mockResolvedValue({ result: AttestationResult.FailureUserConsent } as any);

        renderDialog();

        fireEvent.change(getDescription(), { target: { value: "My Key" } });
        clickNext();

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith("You cancelled the attestation request"),
        );
    });

    it("reports an excluded credential", async () => {
        startRegistrationMock.mockResolvedValue({ result: AttestationResult.FailureExcluded } as any);

        renderDialog();

        fireEvent.change(getDescription(), { target: { value: "My Key" } });
        clickNext();

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith("You have registered this device already"),
        );
    });

    it("reports a timed out registration", async () => {
        startRegistrationMock.mockRejectedValue(new Error("timeout"));

        renderDialog();

        fireEvent.change(getDescription(), { target: { value: "My Key" } });
        clickNext();

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith(
                "Failed to register your credential, the identity verification process might have timed out",
            ),
        );
        expect(console.error).toHaveBeenCalled();
    });

    it("shows the touch prompt while waiting", async () => {
        let resolve: (value: unknown) => void = () => {};
        startRegistrationMock.mockReturnValue(new Promise((r) => (resolve = r)) as any);

        renderDialog();

        fireEvent.change(getDescription(), { target: { value: "My Key" } });
        clickNext();

        expect(await screen.findByText("Touch the token on your security key")).toBeInTheDocument();
        expect(screen.getByTestId("webauthn-register-icon")).toHaveAttribute("data-timeout", "60000");
        expect(document.getElementById("dialog-cancel")).toBeDisabled();
        expect(document.getElementById("dialog-next")).toBeNull();

        resolve({ response: { id: "credential-id" }, result: AttestationResult.Success });
        await waitFor(() => expect(finishRegistrationMock).toHaveBeenCalled());
    });

    it("omits the register icon when the options carry no timeout", async () => {
        getOptionsMock.mockResolvedValue({ options: { challenge: "challenge" }, status: 200 } as any);
        let resolve: (value: unknown) => void = () => {};
        startRegistrationMock.mockReturnValue(new Promise((r) => (resolve = r)) as any);

        renderDialog();

        fireEvent.change(getDescription(), { target: { value: "My Key" } });
        clickNext();

        expect(await screen.findByText("Touch the token on your security key")).toBeInTheDocument();
        expect(screen.queryByTestId("webauthn-register-icon")).not.toBeInTheDocument();

        resolve({ response: { id: "credential-id" }, result: AttestationResult.Success });
        await waitFor(() => expect(finishRegistrationMock).toHaveBeenCalled());
    });
});

describe("closing", () => {
    it("closes and resets when cancelled on the first step", async () => {
        const setClosed = vi.fn();

        renderDialog({ setClosed });

        fireEvent.change(getDescription(), { target: { value: "My Key" } });
        clickCancel();

        await waitFor(() => expect(setClosed).toHaveBeenCalled());
        await waitFor(() => expect(getDescription()).toHaveValue(""));
    });
});

describe("dismissal", () => {
    it("closes when Escape is pressed on the first step", async () => {
        const setClosed = vi.fn();

        renderDialog({ setClosed });

        fireEvent.keyDown(document.activeElement ?? document.body, { key: "Escape" });

        await waitFor(() => expect(setClosed).toHaveBeenCalled());
    });

    it("ignores Escape while waiting for the security key", async () => {
        let resolve: (value: unknown) => void = () => {};
        startRegistrationMock.mockReturnValue(new Promise((r) => (resolve = r)) as any);

        const setClosed = vi.fn();
        renderDialog({ setClosed });

        fireEvent.change(getDescription(), { target: { value: "My Key" } });
        clickNext();
        await screen.findByText("Touch the token on your security key");

        fireEvent.keyDown(document.activeElement ?? document.body, { key: "Escape" });

        expect(setClosed).not.toHaveBeenCalled();

        resolve({ response: { id: "credential-id" }, result: AttestationResult.Success });
        await waitFor(() => expect(finishRegistrationMock).toHaveBeenCalled());
    });
});
