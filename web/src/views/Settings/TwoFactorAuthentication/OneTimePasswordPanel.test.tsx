// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import { getUserSessionElevation } from "@services/UserSessionElevation";
import OneTimePasswordPanel from "@views/Settings/TwoFactorAuthentication/OneTimePasswordPanel";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@services/UserSessionElevation", () => ({
    getUserSessionElevation: vi.fn(),
}));

vi.mock("@views/Settings/Common/IdentityVerificationDialog", () => ({
    default: (props: any) => (
        <div data-testid="identity-dialog" data-opening={String(props.opening)}>
            <button data-testid="iv-closed-ok" onClick={() => props.handleClosed(true)} />
            <button data-testid="iv-closed-cancel" onClick={() => props.handleClosed(false)} />
            <button data-testid="iv-opened" onClick={() => props.handleOpened()} />
        </div>
    ),
}));

vi.mock("@views/Settings/Common/SecondFactorDialog", () => ({
    default: (props: any) => (
        <div
            data-testid="second-factor-dialog"
            data-opening={String(props.opening)}
            data-elevation={JSON.stringify(props.elevation ?? null)}
        >
            <button data-testid="sf-closed-ok-changed" onClick={() => props.handleClosed(true, true)} />
            <button data-testid="sf-closed-ok-unchanged" onClick={() => props.handleClosed(true, false)} />
            <button data-testid="sf-closed-cancel" onClick={() => props.handleClosed(false, false)} />
            <button data-testid="sf-opened" onClick={() => props.handleOpened()} />
        </div>
    ),
}));

vi.mock("@views/Settings/TwoFactorAuthentication/OneTimePasswordConfiguration", () => ({
    default: (props: any) => (
        <div data-testid="otp-config" data-issuer={props.config?.issuer}>
            <button data-testid="config-information" onClick={() => props.handleInformation()} />
            <button data-testid="config-delete" onClick={() => props.handleDelete()} />
        </div>
    ),
}));

vi.mock("@views/Settings/TwoFactorAuthentication/OneTimePasswordDeleteDialog", () => ({
    default: (props: any) => (
        <div data-testid="otp-delete-dialog" data-open={String(props.open)}>
            <button data-testid="delete-close" onClick={() => props.handleClose()} />
        </div>
    ),
}));

vi.mock("@views/Settings/TwoFactorAuthentication/OneTimePasswordInformationDialog", () => ({
    default: (props: any) => (
        <div data-testid="otp-info-dialog" data-open={String(props.open)} data-issuer={props.config?.issuer}>
            <button data-testid="info-close" onClick={() => props.handleClose()} />
        </div>
    ),
}));

vi.mock("@views/Settings/TwoFactorAuthentication/OneTimePasswordRegisterDialog", () => ({
    default: (props: any) => (
        <div data-testid="otp-register-dialog" data-open={String(props.open)}>
            <button data-testid="register-close" onClick={() => props.setClosed()} />
        </div>
    ),
}));

const getElevationMock = vi.mocked(getUserSessionElevation);

const config = { digits: 6, issuer: "Authelia", period: 30 } as any;

const elevated = { elevated: true, skip_second_factor: false } as any;
const notElevated = { elevated: false, skip_second_factor: false } as any;
const skipSecondFactor = { elevated: false, skip_second_factor: true } as any;

function renderPanel(props: Partial<{ config: any; handleRefreshState: () => void; info: any }> = {}) {
    const handleRefreshState = props.handleRefreshState ?? vi.fn();
    const merged = { config: null, info: undefined, ...props, handleRefreshState };

    return { ...render(<OneTimePasswordPanel {...merged} />), handleRefreshState };
}

function getAdd() {
    return document.getElementById("one-time-password-add") as HTMLButtonElement;
}

beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, "warn").mockImplementation(() => {});
    vi.spyOn(console, "error").mockImplementation(() => {});
    getElevationMock.mockResolvedValue(elevated);
});

afterEach(() => {
    vi.restoreAllMocks();
});

describe("rendering", () => {
    it("renders panel with title and add button", () => {
        renderPanel({ config: undefined });
        expect(screen.getByText("One-Time Password")).toBeInTheDocument();
        expect(screen.getByText("Add")).toBeInTheDocument();
    });

    it("does not claim the One-Time Password is unregistered while it is still loading", () => {
        const { container } = renderPanel({ config: undefined });
        expect(
            screen.queryByText("The One-Time Password has not been registered if you'd like to register it click add"),
        ).not.toBeInTheDocument();
        expect(container.querySelector('[data-loading="true"]')).toBeInTheDocument();
    });

    it("renders not registered message when config is null", () => {
        renderPanel({ config: null });
        expect(
            screen.getByText("The One-Time Password has not been registered if you'd like to register it click add"),
        ).toBeInTheDocument();
        expect(document.querySelector('[data-loading="false"]')).toBeInTheDocument();
    });

    it("renders OTP configuration when config is provided", () => {
        renderPanel({ config });
        expect(screen.getByTestId("otp-config")).toHaveAttribute("data-issuer", "Authelia");
    });

    it("disables add button when config is provided", () => {
        renderPanel({ config });
        expect(getAdd()).toBeDisabled();
    });

    it("disables add button while the configuration is still loading", () => {
        renderPanel({ config: undefined });
        expect(getAdd()).toBeDisabled();
    });

    it("enables add button when nothing is registered", () => {
        renderPanel({ config: null });
        expect(getAdd()).not.toBeDisabled();
    });

    it("passes the configuration to the information dialog", () => {
        renderPanel({ config });
        expect(screen.getByTestId("otp-info-dialog")).toHaveAttribute("data-issuer", "Authelia");
    });
});

describe("registration flow", () => {
    it("requests elevation when adding", async () => {
        renderPanel({ config: null });

        fireEvent.click(getAdd());

        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());
        expect(screen.getByTestId("second-factor-dialog")).toHaveAttribute("data-opening", "true");
    });

    it("disables the add button while opening", async () => {
        renderPanel({ config: null });

        fireEvent.click(getAdd());

        await waitFor(() => expect(getAdd()).toBeDisabled());
    });

    it("logs elevation lookup failures", async () => {
        getElevationMock.mockRejectedValue(new Error("boom"));

        renderPanel({ config: null });

        fireEvent.click(getAdd());

        await waitFor(() => expect(console.error).toHaveBeenCalled());
    });

    it("opens the register dialog when already elevated", async () => {
        renderPanel({ config: null });

        fireEvent.click(getAdd());
        await waitFor(() =>
            expect(screen.getByTestId("second-factor-dialog")).toHaveAttribute(
                "data-elevation",
                JSON.stringify(elevated),
            ),
        );

        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));

        await waitFor(() => expect(screen.getByTestId("otp-register-dialog")).toHaveAttribute("data-open", "true"));
    });

    it("treats skip_second_factor as elevated", async () => {
        getElevationMock.mockResolvedValue(skipSecondFactor);

        renderPanel({ config: null });

        fireEvent.click(getAdd());
        await waitFor(() =>
            expect(screen.getByTestId("second-factor-dialog")).toHaveAttribute(
                "data-elevation",
                JSON.stringify(skipSecondFactor),
            ),
        );

        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));

        await waitFor(() => expect(screen.getByTestId("otp-register-dialog")).toHaveAttribute("data-open", "true"));
    });

    it("asks for identity verification when not elevated", async () => {
        getElevationMock.mockResolvedValue(notElevated);

        renderPanel({ config: null });

        fireEvent.click(getAdd());
        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());

        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));

        await waitFor(() => expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "true"));
    });

    it("re-checks elevation when the second factor dialog reports a change", async () => {
        getElevationMock.mockResolvedValueOnce(notElevated).mockResolvedValueOnce(elevated);

        renderPanel({ config: null });

        fireEvent.click(getAdd());
        await waitFor(() => expect(getElevationMock).toHaveBeenCalledTimes(1));

        fireEvent.click(screen.getByTestId("sf-closed-ok-changed"));

        await waitFor(() => expect(getElevationMock).toHaveBeenCalledTimes(2));
        await waitFor(() => expect(screen.getByTestId("otp-register-dialog")).toHaveAttribute("data-open", "true"));
    });

    it("asks for identity verification when the refreshed elevation is insufficient", async () => {
        getElevationMock.mockResolvedValue(notElevated);

        renderPanel({ config: null });

        fireEvent.click(getAdd());
        await waitFor(() => expect(getElevationMock).toHaveBeenCalledTimes(1));

        fireEvent.click(screen.getByTestId("sf-closed-ok-changed"));

        await waitFor(() => expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "true"));
    });

    it("does nothing when the elevation refresh fails after a change", async () => {
        getElevationMock.mockResolvedValueOnce(elevated).mockRejectedValueOnce(new Error("boom"));

        renderPanel({ config: null });

        fireEvent.click(getAdd());
        await waitFor(() => expect(getElevationMock).toHaveBeenCalledTimes(1));

        fireEvent.click(screen.getByTestId("sf-closed-ok-changed"));

        await waitFor(() => expect(console.error).toHaveBeenCalled());
        expect(screen.getByTestId("otp-register-dialog")).toHaveAttribute("data-open", "false");
    });

    it("resets the state when the second factor dialog is cancelled", async () => {
        renderPanel({ config: null });

        fireEvent.click(getAdd());
        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());

        fireEvent.click(screen.getByTestId("sf-closed-cancel"));

        await waitFor(() => expect(getAdd()).not.toBeDisabled());
        expect(console.warn).toHaveBeenCalled();
    });

    it("clears the second factor opening flag once it opens", async () => {
        renderPanel({ config: null });

        fireEvent.click(getAdd());
        await waitFor(() => expect(screen.getByTestId("second-factor-dialog")).toHaveAttribute("data-opening", "true"));

        fireEvent.click(screen.getByTestId("sf-opened"));

        await waitFor(() =>
            expect(screen.getByTestId("second-factor-dialog")).toHaveAttribute("data-opening", "false"),
        );
    });

    it("opens the register dialog once identity is verified", async () => {
        getElevationMock.mockResolvedValue(notElevated);

        renderPanel({ config: null });

        fireEvent.click(getAdd());
        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());
        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));
        await waitFor(() => expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "true"));

        fireEvent.click(screen.getByTestId("iv-closed-ok"));

        await waitFor(() => expect(screen.getByTestId("otp-register-dialog")).toHaveAttribute("data-open", "true"));
    });

    it("resets the state when identity verification is cancelled", async () => {
        getElevationMock.mockResolvedValue(notElevated);

        renderPanel({ config: null });

        fireEvent.click(getAdd());
        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());
        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));
        await waitFor(() => expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "true"));

        fireEvent.click(screen.getByTestId("iv-closed-cancel"));

        await waitFor(() => expect(screen.getByTestId("otp-register-dialog")).toHaveAttribute("data-open", "false"));
        expect(console.warn).toHaveBeenCalled();
    });

    it("clears the identity verification opening flag once it opens", async () => {
        getElevationMock.mockResolvedValue(notElevated);

        renderPanel({ config: null });

        fireEvent.click(getAdd());
        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());
        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));
        await waitFor(() => expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "true"));

        fireEvent.click(screen.getByTestId("iv-opened"));

        await waitFor(() => expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "false"));
    });

    it("refreshes the parent state when the register dialog closes", async () => {
        const { handleRefreshState } = renderPanel({ config: null });

        fireEvent.click(screen.getByTestId("register-close"));

        await waitFor(() => expect(handleRefreshState).toHaveBeenCalled());
    });
});

describe("deletion flow", () => {
    it("requests elevation when deleting a registered configuration", async () => {
        renderPanel({ config });

        fireEvent.click(screen.getByTestId("config-delete"));

        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());
        expect(screen.getByTestId("second-factor-dialog")).toHaveAttribute("data-opening", "true");
    });

    it("opens the delete dialog when already elevated", async () => {
        renderPanel({ config });

        fireEvent.click(screen.getByTestId("config-delete"));
        await waitFor(() =>
            expect(screen.getByTestId("second-factor-dialog")).toHaveAttribute(
                "data-elevation",
                JSON.stringify(elevated),
            ),
        );

        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));

        await waitFor(() => expect(screen.getByTestId("otp-delete-dialog")).toHaveAttribute("data-open", "true"));
    });

    it("opens the delete dialog once identity is verified", async () => {
        getElevationMock.mockResolvedValue(notElevated);

        renderPanel({ config });

        fireEvent.click(screen.getByTestId("config-delete"));
        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());
        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));
        await waitFor(() => expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "true"));

        fireEvent.click(screen.getByTestId("iv-closed-ok"));

        await waitFor(() => expect(screen.getByTestId("otp-delete-dialog")).toHaveAttribute("data-open", "true"));
    });

    it("opens the delete dialog after a re-checked elevation", async () => {
        getElevationMock.mockResolvedValueOnce(notElevated).mockResolvedValueOnce(elevated);

        renderPanel({ config });

        fireEvent.click(screen.getByTestId("config-delete"));
        await waitFor(() => expect(getElevationMock).toHaveBeenCalledTimes(1));

        fireEvent.click(screen.getByTestId("sf-closed-ok-changed"));

        await waitFor(() => expect(screen.getByTestId("otp-delete-dialog")).toHaveAttribute("data-open", "true"));
    });

    it("refreshes the parent state when the delete dialog closes", async () => {
        const { handleRefreshState } = renderPanel({ config });

        fireEvent.click(screen.getByTestId("delete-close"));

        await waitFor(() => expect(handleRefreshState).toHaveBeenCalled());
    });
});

describe("information dialog", () => {
    it("opens without requiring elevation", async () => {
        renderPanel({ config });

        fireEvent.click(screen.getByTestId("config-information"));

        await waitFor(() => expect(screen.getByTestId("otp-info-dialog")).toHaveAttribute("data-open", "true"));
        expect(getElevationMock).not.toHaveBeenCalled();
    });

    it("closes without refreshing the parent state", async () => {
        const { handleRefreshState } = renderPanel({ config });

        fireEvent.click(screen.getByTestId("config-information"));
        await waitFor(() => expect(screen.getByTestId("otp-info-dialog")).toHaveAttribute("data-open", "true"));

        fireEvent.click(screen.getByTestId("info-close"));

        await waitFor(() => expect(screen.getByTestId("otp-info-dialog")).toHaveAttribute("data-open", "false"));
        expect(handleRefreshState).not.toHaveBeenCalled();
    });
});
