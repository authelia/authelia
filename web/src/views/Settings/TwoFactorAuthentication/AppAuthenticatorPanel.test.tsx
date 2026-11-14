// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import { deleteAppAuthenticatorDevice } from "@services/AppAuthenticator";
import { getUserSessionElevation } from "@services/UserSessionElevation";
import AppAuthenticatorPanel from "@views/Settings/TwoFactorAuthentication/AppAuthenticatorPanel";

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

vi.mock("@services/AppAuthenticator", () => ({
    deleteAppAuthenticatorDevice: vi.fn(),
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

vi.mock("@views/Settings/TwoFactorAuthentication/DeleteDialog", () => ({
    default: (props: any) => (
        <div data-testid="delete-dialog" data-open={String(props.open)} data-text={props.text}>
            <button data-testid="delete-confirm" onClick={() => props.onConfirm()} />
            <button data-testid="delete-cancel" onClick={() => props.onCancel()} />
        </div>
    ),
}));

vi.mock("@views/Settings/TwoFactorAuthentication/AppAuthenticatorDeviceEditDialog", () => ({
    default: (props: any) => (
        <div
            data-testid="edit-dialog"
            data-open={String(props.open)}
            data-device={JSON.stringify(props.device ?? null)}
        >
            <button data-testid="edit-close" onClick={() => props.handleClose()} />
        </div>
    ),
}));

vi.mock("@views/Settings/TwoFactorAuthentication/AppAuthenticatorRegisterDialog", () => ({
    default: (props: any) => (
        <div data-testid="register-dialog" data-open={String(props.open)}>
            <button data-testid="register-close" onClick={() => props.setClosed()} />
        </div>
    ),
}));

vi.mock("@views/Settings/TwoFactorAuthentication/AppAuthenticatorDeviceItem", () => ({
    default: (props: any) => (
        <div data-testid={`device-${props.index}`}>
            {props.device.name}
            <button data-testid={`device-${props.index}-edit`} onClick={() => props.handleEdit(props.index)} />
            <button data-testid={`device-${props.index}-delete`} onClick={() => props.handleDelete(props.index)} />
        </div>
    ),
}));

const getElevationMock = vi.mocked(getUserSessionElevation);
const deleteMock = vi.mocked(deleteAppAuthenticatorDevice);

const devices = [
    { id: 11, name: "Jane's iPhone" },
    { id: 12, name: "Work phone" },
] as any;

const elevated = { elevated: true, skip_second_factor: false } as any;
const notElevated = { elevated: false, skip_second_factor: false } as any;

function renderPanel(props: Partial<{ devices: any; registrationAvailable: boolean }> = {}) {
    const handleRefreshState = vi.fn();
    const merged = { devices, info: undefined, registrationAvailable: true, ...props, handleRefreshState };

    return { ...render(<AppAuthenticatorPanel {...merged} />), handleRefreshState };
}

beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, "warn").mockImplementation(() => {});
    vi.spyOn(console, "error").mockImplementation(() => {});
    getElevationMock.mockResolvedValue(elevated);
    deleteMock.mockResolvedValue({ data: { status: "OK" } } as any);
});

afterEach(() => {
    vi.restoreAllMocks();
});

describe("rendering", () => {
    it("renders the title, add button and devices", () => {
        renderPanel();

        expect(screen.getByText("Authelia app")).toBeInTheDocument();
        expect(screen.getByText("Add")).toBeInTheDocument();
        expect(screen.getByTestId("device-0")).toHaveTextContent("Jane's iPhone");
        expect(screen.getByTestId("device-1")).toHaveTextContent("Work phone");
    });

    it("hides the add button when registration is unavailable", () => {
        renderPanel({ registrationAvailable: false });

        expect(screen.queryByText("Add")).not.toBeInTheDocument();
        expect(
            screen.getByText("Registration of the Authelia app is not configured on this server"),
        ).toBeInTheDocument();
    });

    it("shows the empty message when there are no devices", () => {
        renderPanel({ devices: [] });

        expect(
            screen.getByText("No Authelia app devices have been registered if you'd like to register one click add"),
        ).toBeInTheDocument();
    });

    it("does not show the empty message while loading", () => {
        renderPanel({ devices: undefined });

        expect(
            screen.queryByText("No Authelia app devices have been registered if you'd like to register one click add"),
        ).not.toBeInTheDocument();
    });
});

describe("register", () => {
    it("opens the register dialog after elevation and refreshes when it closes", async () => {
        const { handleRefreshState } = renderPanel();

        fireEvent.click(screen.getByText("Add"));

        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());
        expect(screen.getByTestId("register-dialog")).toHaveAttribute("data-open", "false");

        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));

        expect(screen.getByTestId("register-dialog")).toHaveAttribute("data-open", "true");

        fireEvent.click(screen.getByTestId("register-close"));

        expect(screen.getByTestId("register-dialog")).toHaveAttribute("data-open", "false");
        expect(handleRefreshState).toHaveBeenCalled();
    });
});

describe("delete", () => {
    it("deletes the device only after elevation", async () => {
        const { handleRefreshState } = renderPanel();

        fireEvent.click(screen.getByTestId("device-1-delete"));

        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());
        expect(screen.getByTestId("delete-dialog")).toHaveAttribute("data-open", "false");
        expect(deleteMock).not.toHaveBeenCalled();

        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));

        expect(screen.getByTestId("delete-dialog")).toHaveAttribute("data-open", "true");

        fireEvent.click(screen.getByTestId("delete-confirm"));

        await waitFor(() => expect(deleteMock).toHaveBeenCalledWith(12));
        await waitFor(() => expect(handleRefreshState).toHaveBeenCalled());
        expect(mocks.createSuccessNotification).toHaveBeenCalled();
        expect(screen.getByTestId("delete-dialog")).toHaveAttribute("data-open", "false");
    });

    it("deletes the device after identity verification", async () => {
        getElevationMock.mockResolvedValue(notElevated);

        renderPanel();

        fireEvent.click(screen.getByTestId("device-0-delete"));

        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());
        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));

        expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "true");

        fireEvent.click(screen.getByTestId("iv-closed-ok"));
        fireEvent.click(screen.getByTestId("delete-confirm"));

        await waitFor(() => expect(deleteMock).toHaveBeenCalledWith(11));
    });

    it("does not delete when elevation is cancelled", async () => {
        renderPanel();

        fireEvent.click(screen.getByTestId("device-0-delete"));

        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());
        fireEvent.click(screen.getByTestId("sf-closed-cancel"));

        expect(screen.getByTestId("delete-dialog")).toHaveAttribute("data-open", "false");
        expect(deleteMock).not.toHaveBeenCalled();
    });

    it("reports an elevation error from the endpoint", async () => {
        deleteMock.mockResolvedValue({ data: { elevation: true, status: "KO" } } as any);

        const { handleRefreshState } = renderPanel();

        fireEvent.click(screen.getByTestId("device-0-delete"));
        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());
        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));
        fireEvent.click(screen.getByTestId("delete-confirm"));

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith("You must be elevated to {{action}} a {{item}}"),
        );
        expect(handleRefreshState).not.toHaveBeenCalled();
    });

    it("reports a failed request", async () => {
        deleteMock.mockRejectedValue(new Error("not found"));

        renderPanel();

        fireEvent.click(screen.getByTestId("device-0-delete"));
        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());
        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));
        fireEvent.click(screen.getByTestId("delete-confirm"));

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith("There was a problem {{action}} the {{item}}"),
        );
    });
});

describe("edit", () => {
    it("opens the edit dialog with the device after elevation", async () => {
        const { handleRefreshState } = renderPanel();

        fireEvent.click(screen.getByTestId("device-1-edit"));

        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());
        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));

        expect(screen.getByTestId("edit-dialog")).toHaveAttribute("data-open", "true");
        expect(screen.getByTestId("edit-dialog")).toHaveAttribute("data-device", JSON.stringify(devices[1]));

        fireEvent.click(screen.getByTestId("edit-close"));

        expect(screen.getByTestId("edit-dialog")).toHaveAttribute("data-open", "false");
        expect(handleRefreshState).toHaveBeenCalled();
    });
});

describe("elevation", () => {
    it("refreshes the elevation after a changed second factor result", async () => {
        renderPanel();

        fireEvent.click(screen.getByTestId("device-0-delete"));
        await waitFor(() => expect(getElevationMock).toHaveBeenCalledTimes(1));

        fireEvent.click(screen.getByTestId("sf-closed-ok-changed"));

        await waitFor(() => expect(screen.getByTestId("delete-dialog")).toHaveAttribute("data-open", "true"));
        expect(getElevationMock).toHaveBeenCalledTimes(2);
    });

    it("asks for identity verification when a changed result is still not elevated", async () => {
        getElevationMock.mockResolvedValue(notElevated);

        renderPanel();

        fireEvent.click(screen.getByTestId("device-0-delete"));
        await waitFor(() => expect(getElevationMock).toHaveBeenCalledTimes(1));

        fireEvent.click(screen.getByTestId("sf-closed-ok-changed"));

        await waitFor(() => expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "true"));
    });

    it("resets when identity verification is cancelled", async () => {
        getElevationMock.mockResolvedValue(notElevated);

        renderPanel();

        fireEvent.click(screen.getByTestId("device-0-delete"));
        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());
        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));
        fireEvent.click(screen.getByTestId("iv-closed-cancel"));

        expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "false");
        expect(screen.getByTestId("delete-dialog")).toHaveAttribute("data-open", "false");
    });

    it("marks the dialogs as opened", async () => {
        renderPanel();

        fireEvent.click(screen.getByTestId("device-0-delete"));
        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());

        fireEvent.click(screen.getByTestId("sf-opened"));
        fireEvent.click(screen.getByTestId("iv-opened"));

        expect(screen.getByTestId("second-factor-dialog")).toHaveAttribute("data-opening", "false");
        expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "false");
    });
});

describe("delete responses", () => {
    it("reports a KO response requiring a higher authentication level", async () => {
        deleteMock.mockResolvedValue({ data: { authentication: true, status: "KO" } } as any);

        renderPanel();

        fireEvent.click(screen.getByTestId("device-0-delete"));
        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());
        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));
        fireEvent.click(screen.getByTestId("delete-confirm"));

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith(
                "You must have a higher authentication level to {{action}} a {{item}}",
            ),
        );
    });

    it("reports any other KO response", async () => {
        deleteMock.mockResolvedValue({ data: { status: "KO" } } as any);

        renderPanel();

        fireEvent.click(screen.getByTestId("device-0-delete"));
        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());
        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));
        fireEvent.click(screen.getByTestId("delete-confirm"));

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith("There was a problem {{action}} the {{item}}"),
        );
    });

    it("closes the delete dialog on cancel without deleting", async () => {
        renderPanel();

        fireEvent.click(screen.getByTestId("device-0-delete"));
        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());
        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));
        fireEvent.click(screen.getByTestId("delete-cancel"));

        expect(screen.getByTestId("delete-dialog")).toHaveAttribute("data-open", "false");
        expect(deleteMock).not.toHaveBeenCalled();
    });
});
