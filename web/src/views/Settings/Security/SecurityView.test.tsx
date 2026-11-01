// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import { getUserSessionElevation } from "@services/UserSessionElevation";
import SecurityView from "@views/Settings/Security/SecurityView";

const mocks = vi.hoisted(() => ({
    configuration: { password_change_disabled: false } as any,
    configurationError: null as Error | null,
    createErrorNotification: vi.fn(),
    createSuccessNotification: vi.fn(),
    fetchConfiguration: vi.fn(),
    fetchUserInfo: vi.fn(),
    userInfo: { display_name: "John Doe", emails: ["john@example.com"], groups: [] } as any,
    userInfoError: null as Error | null,
}));

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@hooks/Configuration", () => ({
    useConfiguration: () => [mocks.configuration, mocks.fetchConfiguration, false, mocks.configurationError],
}));

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: mocks.createErrorNotification,
        createSuccessNotification: mocks.createSuccessNotification,
    }),
}));

vi.mock("@hooks/UserInfo", () => ({
    useUserInfoGET: () => [mocks.userInfo, mocks.fetchUserInfo, false, mocks.userInfoError],
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
            data-display-name={props.info?.display_name ?? ""}
        >
            <button data-testid="sf-closed-ok-changed" onClick={() => props.handleClosed(true, true)} />
            <button data-testid="sf-closed-ok-unchanged" onClick={() => props.handleClosed(true, false)} />
            <button data-testid="sf-closed-cancel" onClick={() => props.handleClosed(false, false)} />
            <button data-testid="sf-opened" onClick={() => props.handleOpened()} />
        </div>
    ),
}));

vi.mock("@views/Settings/Security/ChangePasswordDialog", () => ({
    default: (props: any) => (
        <div data-testid="change-password-dialog" data-open={String(props.open)} data-username={props.username}>
            <button data-testid="pw-close" onClick={() => props.setClosed()} />
        </div>
    ),
}));

const getElevationMock = vi.mocked(getUserSessionElevation);

const elevated = { elevated: true, skip_second_factor: false } as any;
const notElevated = { elevated: false, skip_second_factor: false } as any;
const skipSecondFactor = { elevated: false, skip_second_factor: true } as any;

function getChangePasswordButton() {
    return document.getElementById("change-password-button") as HTMLButtonElement;
}

beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, "warn").mockImplementation(() => {});
    vi.spyOn(console, "error").mockImplementation(() => {});
    mocks.configuration = { password_change_disabled: false };
    mocks.configurationError = null;
    mocks.userInfo = { display_name: "John Doe", emails: ["john@example.com"], groups: [] };
    mocks.userInfoError = null;
    getElevationMock.mockResolvedValue(elevated);
});

afterEach(() => {
    vi.restoreAllMocks();
});

describe("rendering", () => {
    it("renders user info and change password button", () => {
        render(<SecurityView />);
        expect(screen.getByText(/John Doe/)).toBeInTheDocument();
        expect(screen.getByText("Change Password")).toBeInTheDocument();
    });

    it("renders dialogs", () => {
        render(<SecurityView />);
        expect(screen.getByTestId("identity-dialog")).toBeInTheDocument();
        expect(screen.getByTestId("second-factor-dialog")).toBeInTheDocument();
        expect(screen.getByTestId("change-password-dialog")).toBeInTheDocument();
    });

    it("fetches the user info and configuration on mount", () => {
        render(<SecurityView />);
        expect(mocks.fetchUserInfo).toHaveBeenCalled();
        expect(mocks.fetchConfiguration).toHaveBeenCalled();
    });

    it("shows the primary email address", () => {
        render(<SecurityView />);
        expect(screen.getByText("john@example.com")).toBeInTheDocument();
    });

    it("lists additional email addresses", () => {
        mocks.userInfo = {
            display_name: "John Doe",
            emails: ["john@example.com", "j.doe@example.com", "jd@example.com"],
            groups: [],
        };

        render(<SecurityView />);

        expect(screen.getByText("j.doe@example.com")).toBeInTheDocument();
        expect(screen.getByText("jd@example.com")).toBeInTheDocument();
    });

    it("does not list additional emails when only one exists", () => {
        render(<SecurityView />);
        expect(document.querySelector("ul")).toBeNull();
    });

    it("tolerates missing user info", () => {
        mocks.userInfo = undefined;

        render(<SecurityView />);

        expect(screen.getByText("Change Password")).toBeInTheDocument();
        expect(screen.getByTestId("change-password-dialog")).toHaveAttribute("data-username", "");
    });

    it("tolerates user info without email addresses", () => {
        mocks.userInfo = { display_name: "John Doe", groups: [] };

        render(<SecurityView />);

        expect(screen.getByText("Email:")).toBeInTheDocument();
    });

    it("passes the display name to the dialogs", () => {
        render(<SecurityView />);
        expect(screen.getByTestId("change-password-dialog")).toHaveAttribute("data-username", "John Doe");
        expect(screen.getByTestId("second-factor-dialog")).toHaveAttribute("data-display-name", "John Doe");
    });
});

describe("configuration gating", () => {
    it("enables the button when password change is allowed", () => {
        render(<SecurityView />);
        expect(getChangePasswordButton()).not.toBeDisabled();
    });

    it("disables the button when the administrator disabled password change", () => {
        mocks.configuration = { password_change_disabled: true };

        render(<SecurityView />);

        expect(getChangePasswordButton()).toBeDisabled();
    });

    it("disables the button while the configuration is unavailable", () => {
        mocks.configuration = undefined;

        render(<SecurityView />);

        expect(getChangePasswordButton()).toBeDisabled();
    });
});

describe("fetch errors", () => {
    it("notifies when the user info cannot be retrieved", () => {
        mocks.userInfoError = new Error("boom");

        render(<SecurityView />);

        expect(mocks.createErrorNotification).toHaveBeenCalledWith("There was an issue retrieving user preferences");
    });

    it("notifies when the configuration cannot be retrieved", () => {
        mocks.configurationError = new Error("boom");

        render(<SecurityView />);

        expect(mocks.createErrorNotification).toHaveBeenCalledWith("There was an issue retrieving configuration");
    });
});

describe("password change flow", () => {
    it("requests elevation when the button is pressed", async () => {
        render(<SecurityView />);

        fireEvent.click(getChangePasswordButton());

        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());
        expect(screen.getByTestId("second-factor-dialog")).toHaveAttribute("data-opening", "true");
    });

    it("logs elevation lookup failures", async () => {
        getElevationMock.mockRejectedValue(new Error("boom"));

        render(<SecurityView />);

        fireEvent.click(getChangePasswordButton());

        await waitFor(() => expect(console.error).toHaveBeenCalled());
    });

    it("opens the change password dialog when already elevated", async () => {
        render(<SecurityView />);

        fireEvent.click(getChangePasswordButton());
        await waitFor(() =>
            expect(screen.getByTestId("second-factor-dialog")).toHaveAttribute(
                "data-elevation",
                JSON.stringify(elevated),
            ),
        );

        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));

        await waitFor(() => expect(screen.getByTestId("change-password-dialog")).toHaveAttribute("data-open", "true"));
    });

    it("treats skip_second_factor as elevated", async () => {
        getElevationMock.mockResolvedValue(skipSecondFactor);

        render(<SecurityView />);

        fireEvent.click(getChangePasswordButton());
        await waitFor(() =>
            expect(screen.getByTestId("second-factor-dialog")).toHaveAttribute(
                "data-elevation",
                JSON.stringify(skipSecondFactor),
            ),
        );

        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));

        await waitFor(() => expect(screen.getByTestId("change-password-dialog")).toHaveAttribute("data-open", "true"));
    });

    it("asks for identity verification when not elevated", async () => {
        getElevationMock.mockResolvedValue(notElevated);

        render(<SecurityView />);

        fireEvent.click(getChangePasswordButton());
        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());

        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));

        await waitFor(() => expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "true"));
    });

    it("re-checks elevation when the dialog reports a change", async () => {
        getElevationMock.mockResolvedValueOnce(notElevated).mockResolvedValueOnce(elevated);

        render(<SecurityView />);

        fireEvent.click(getChangePasswordButton());
        await waitFor(() => expect(getElevationMock).toHaveBeenCalledTimes(1));

        fireEvent.click(screen.getByTestId("sf-closed-ok-changed"));

        await waitFor(() => expect(getElevationMock).toHaveBeenCalledTimes(2));
        await waitFor(() => expect(screen.getByTestId("change-password-dialog")).toHaveAttribute("data-open", "true"));
    });

    it("asks for identity verification when the refreshed elevation is insufficient", async () => {
        getElevationMock.mockResolvedValue(notElevated);

        render(<SecurityView />);

        fireEvent.click(getChangePasswordButton());
        await waitFor(() => expect(getElevationMock).toHaveBeenCalledTimes(1));

        fireEvent.click(screen.getByTestId("sf-closed-ok-changed"));

        await waitFor(() => expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "true"));
    });

    it("notifies when the elevation refresh fails after a change", async () => {
        getElevationMock.mockResolvedValueOnce(elevated).mockRejectedValueOnce(new Error("boom"));

        render(<SecurityView />);

        fireEvent.click(getChangePasswordButton());
        await waitFor(() => expect(getElevationMock).toHaveBeenCalledTimes(1));

        fireEvent.click(screen.getByTestId("sf-closed-ok-changed"));

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith("Failed to get session elevation status"),
        );
    });

    it("resets the state when the second factor dialog is cancelled", async () => {
        render(<SecurityView />);

        fireEvent.click(getChangePasswordButton());
        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());

        fireEvent.click(screen.getByTestId("sf-closed-cancel"));

        await waitFor(() =>
            expect(screen.getByTestId("second-factor-dialog")).toHaveAttribute("data-opening", "false"),
        );
        expect(console.warn).toHaveBeenCalled();
        expect(screen.getByTestId("change-password-dialog")).toHaveAttribute("data-open", "false");
    });

    it("clears the second factor opening flag once it opens", async () => {
        render(<SecurityView />);

        fireEvent.click(getChangePasswordButton());
        await waitFor(() => expect(screen.getByTestId("second-factor-dialog")).toHaveAttribute("data-opening", "true"));

        fireEvent.click(screen.getByTestId("sf-opened"));

        await waitFor(() =>
            expect(screen.getByTestId("second-factor-dialog")).toHaveAttribute("data-opening", "false"),
        );
    });

    it("opens the change password dialog once identity is verified", async () => {
        getElevationMock.mockResolvedValue(notElevated);

        render(<SecurityView />);

        fireEvent.click(getChangePasswordButton());
        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());
        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));
        await waitFor(() => expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "true"));

        fireEvent.click(screen.getByTestId("iv-closed-ok"));

        await waitFor(() => expect(screen.getByTestId("change-password-dialog")).toHaveAttribute("data-open", "true"));
    });

    it("resets the state when identity verification is cancelled", async () => {
        getElevationMock.mockResolvedValue(notElevated);

        render(<SecurityView />);

        fireEvent.click(getChangePasswordButton());
        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());
        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));
        await waitFor(() => expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "true"));

        fireEvent.click(screen.getByTestId("iv-closed-cancel"));

        await waitFor(() => expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "false"));
        expect(screen.getByTestId("change-password-dialog")).toHaveAttribute("data-open", "false");
    });

    it("clears the identity verification opening flag once it opens", async () => {
        getElevationMock.mockResolvedValue(notElevated);

        render(<SecurityView />);

        fireEvent.click(getChangePasswordButton());
        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());
        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));
        await waitFor(() => expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "true"));

        fireEvent.click(screen.getByTestId("iv-opened"));

        await waitFor(() => expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "false"));
    });

    it("resets the state when the change password dialog closes", async () => {
        render(<SecurityView />);

        fireEvent.click(getChangePasswordButton());
        await waitFor(() =>
            expect(screen.getByTestId("second-factor-dialog")).toHaveAttribute(
                "data-elevation",
                JSON.stringify(elevated),
            ),
        );
        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));
        await waitFor(() => expect(screen.getByTestId("change-password-dialog")).toHaveAttribute("data-open", "true"));

        fireEvent.click(screen.getByTestId("pw-close"));

        await waitFor(() => expect(screen.getByTestId("change-password-dialog")).toHaveAttribute("data-open", "false"));
    });
});
