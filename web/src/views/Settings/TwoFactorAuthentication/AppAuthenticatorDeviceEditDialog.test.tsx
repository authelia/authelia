// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { act, fireEvent, render, screen } from "@testing-library/react";

import { updateAppAuthenticatorDevice } from "@services/AppAuthenticator";
import AppAuthenticatorDeviceEditDialog from "@views/Settings/TwoFactorAuthentication/AppAuthenticatorDeviceEditDialog";

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
    updateAppAuthenticatorDevice: vi.fn(),
}));

const updateMock = vi.mocked(updateAppAuthenticatorDevice);

const device = { id: 5, name: "Jane's iPhone" } as any;

function renderDialog() {
    const handleClose = vi.fn();

    render(<AppAuthenticatorDeviceEditDialog open={true} device={device} handleClose={handleClose} />);

    return handleClose;
}

async function rename(value: string) {
    await act(async () => {
        fireEvent.change(screen.getByLabelText("Name"), { target: { value } });
    });

    await act(async () => {
        fireEvent.click(screen.getByText("Update"));
    });
}

beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, "error").mockImplementation(() => {});
    updateMock.mockResolvedValue({ data: { status: "OK" } } as any);
});

afterEach(() => {
    vi.restoreAllMocks();
});

it("renders the dialog", () => {
    renderDialog();

    expect(screen.getByText("Rename device")).toBeInTheDocument();
});

it("renames the device", async () => {
    const handleClose = renderDialog();

    await rename("  Work phone  ");

    expect(updateMock).toHaveBeenCalledWith(5, "Work phone");
    expect(mocks.createSuccessNotification).toHaveBeenCalled();
    expect(handleClose).toHaveBeenCalledOnce();
});

it("does not rename to an empty name", async () => {
    const handleClose = renderDialog();

    await rename("   ");

    expect(updateMock).not.toHaveBeenCalled();
    expect(handleClose).not.toHaveBeenCalled();
});

it("limits the name to 64 characters", async () => {
    renderDialog();

    await rename("a".repeat(70));

    expect(updateMock).toHaveBeenCalledWith(5, "a".repeat(64));
});

it("reports a KO response", async () => {
    updateMock.mockResolvedValue({ data: { elevation: true, status: "KO" } } as any);

    const handleClose = renderDialog();

    await rename("Work phone");

    expect(mocks.createErrorNotification).toHaveBeenCalledWith("You must be elevated to {{action}} a {{item}}");
    expect(handleClose).toHaveBeenCalledOnce();
});

it("reports a failed request", async () => {
    updateMock.mockRejectedValue(new Error("bad request"));

    renderDialog();

    await rename("Work phone");

    expect(mocks.createErrorNotification).toHaveBeenCalledWith("There was a problem {{action}} the {{item}}");
});

it("closes on cancel", () => {
    const handleClose = renderDialog();

    fireEvent.click(screen.getByText("Cancel"));

    expect(handleClose).toHaveBeenCalledOnce();
    expect(updateMock).not.toHaveBeenCalled();
});

it("reports a KO response requiring a higher authentication level", async () => {
    updateMock.mockResolvedValue({ data: { authentication: true, status: "KO" } } as any);

    renderDialog();

    await rename("Work phone");

    expect(mocks.createErrorNotification).toHaveBeenCalledWith(
        "You must have a higher authentication level to {{action}} a {{item}}",
    );
});

it("reports any other KO response", async () => {
    updateMock.mockResolvedValue({ data: { status: "KO" } } as any);

    renderDialog();

    await rename("Work phone");

    expect(mocks.createErrorNotification).toHaveBeenCalledWith("There was a problem {{action}} the {{item}}");
    expect(mocks.createSuccessNotification).not.toHaveBeenCalled();
});

it("renames the device when enter is pressed", async () => {
    const handleClose = renderDialog();

    await act(async () => {
        fireEvent.change(screen.getByLabelText("Name"), { target: { value: "Work phone" } });
    });

    await act(async () => {
        fireEvent.keyDown(screen.getByLabelText("Name"), { key: "Enter" });
    });

    expect(updateMock).toHaveBeenCalledWith(5, "Work phone");
    expect(handleClose).toHaveBeenCalledOnce();
});

it("flags an empty name when enter is pressed", async () => {
    const handleClose = renderDialog();

    await act(async () => {
        fireEvent.keyDown(screen.getByLabelText("Name"), { key: "Enter" });
    });

    expect(updateMock).not.toHaveBeenCalled();
    expect(handleClose).not.toHaveBeenCalled();
    expect(screen.getByLabelText("Name")).toHaveAttribute("aria-invalid", "true");
});
