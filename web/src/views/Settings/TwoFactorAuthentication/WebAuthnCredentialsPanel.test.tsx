import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import { getUserSessionElevation } from "@services/UserSessionElevation";
import WebAuthnCredentialsPanel from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialsPanel";

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

vi.mock("@views/Settings/TwoFactorAuthentication/WebAuthnCredentialDeleteDialog", () => ({
    default: (props: any) => (
        <div
            data-testid="delete-dialog"
            data-open={String(props.open)}
            data-credential={JSON.stringify(props.credential ?? null)}
        >
            <button data-testid="delete-close" onClick={() => props.handleClose()} />
        </div>
    ),
}));

vi.mock("@views/Settings/TwoFactorAuthentication/WebAuthnCredentialEditDialog", () => ({
    default: (props: any) => (
        <div
            data-testid="edit-dialog"
            data-open={String(props.open)}
            data-credential={JSON.stringify(props.credential ?? null)}
        >
            <button data-testid="edit-close" onClick={() => props.handleClose()} />
        </div>
    ),
}));

vi.mock("@views/Settings/TwoFactorAuthentication/WebAuthnCredentialInformationDialog", () => ({
    default: (props: any) => (
        <div
            data-testid="info-dialog"
            data-open={String(props.open)}
            data-credential={JSON.stringify(props.credential ?? null)}
        >
            <button data-testid="info-close" onClick={() => props.handleClose()} />
        </div>
    ),
}));

vi.mock("@views/Settings/TwoFactorAuthentication/WebAuthnCredentialRegisterDialog", () => ({
    default: (props: any) => (
        <div data-testid="register-dialog" data-open={String(props.open)}>
            <button data-testid="register-close" onClick={() => props.setClosed()} />
        </div>
    ),
}));

vi.mock("@views/Settings/TwoFactorAuthentication/WebAuthnCredentialsGrid", () => ({
    default: (props: any) => (
        <div data-testid="credentials-grid" data-count={props.credentials?.length ?? 0}>
            <button data-testid="grid-information" onClick={() => props.handleInformation(0)} />
            <button data-testid="grid-information-oob" onClick={() => props.handleInformation(9)} />
            <button data-testid="grid-edit" onClick={() => props.handleEdit(0)} />
            <button data-testid="grid-edit-oob" onClick={() => props.handleEdit(9)} />
            <button data-testid="grid-delete" onClick={() => props.handleDelete(0)} />
            <button data-testid="grid-delete-oob" onClick={() => props.handleDelete(9)} />
        </div>
    ),
}));

const getElevationMock = vi.mocked(getUserSessionElevation);

const credentials = [
    { description: "Key 1", id: 1 },
    { description: "Key 2", id: 2 },
] as any;

function renderPanel(props: Partial<{ credentials: any; handleRefreshState: () => void; info: any }> = {}) {
    const handleRefreshState = props.handleRefreshState ?? vi.fn();
    const merged = { credentials, info: undefined, ...props, handleRefreshState };

    return { ...render(<WebAuthnCredentialsPanel {...merged} />), handleRefreshState };
}

const elevated = { elevated: true, skip_second_factor: false } as any;
const notElevated = { elevated: false, skip_second_factor: false } as any;
const skipSecondFactor = { elevated: false, skip_second_factor: true } as any;

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
        renderPanel({ credentials: undefined });
        expect(screen.getByText("WebAuthn Credentials")).toBeInTheDocument();
        expect(screen.getByText("Add")).toBeInTheDocument();
    });

    it("does not claim there are no credentials while they are still loading", () => {
        const { container } = renderPanel({ credentials: undefined });
        expect(
            screen.queryByText("No WebAuthn Credentials have been registered if you'd like to register one click add"),
        ).not.toBeInTheDocument();
        expect(container.querySelector('[data-loading="true"]')).toBeInTheDocument();
    });

    it("renders no credentials message when credentials is empty", () => {
        renderPanel({ credentials: [] });
        expect(
            screen.getByText("No WebAuthn Credentials have been registered if you'd like to register one click add"),
        ).toBeInTheDocument();
    });

    it("renders credentials grid when credentials are provided", () => {
        renderPanel();
        expect(screen.getByTestId("credentials-grid")).toHaveAttribute("data-count", "2");
    });

    it("starts with every dialog closed", () => {
        renderPanel();
        expect(screen.getByTestId("register-dialog")).toHaveAttribute("data-open", "false");
        expect(screen.getByTestId("edit-dialog")).toHaveAttribute("data-open", "false");
        expect(screen.getByTestId("delete-dialog")).toHaveAttribute("data-open", "false");
        expect(screen.getByTestId("info-dialog")).toHaveAttribute("data-open", "false");
    });
});

describe("elevation", () => {
    it("requests elevation when adding a credential", async () => {
        renderPanel();

        fireEvent.click(screen.getByRole("button", { name: /Add/ }));

        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());
        expect(screen.getByTestId("second-factor-dialog")).toHaveAttribute("data-opening", "true");
    });

    it("disables the add button while the register flow is opening", async () => {
        renderPanel();

        fireEvent.click(screen.getByRole("button", { name: /Add/ }));

        await waitFor(() => expect(screen.getByRole("button", { name: /Add/ })).toBeDisabled());
    });

    it("passes the refreshed elevation to the second factor dialog", async () => {
        renderPanel();

        fireEvent.click(screen.getByRole("button", { name: /Add/ }));

        await waitFor(() =>
            expect(screen.getByTestId("second-factor-dialog")).toHaveAttribute(
                "data-elevation",
                JSON.stringify(elevated),
            ),
        );
    });

    it("logs elevation lookup failures", async () => {
        getElevationMock.mockRejectedValue(new Error("boom"));

        renderPanel();

        fireEvent.click(screen.getByRole("button", { name: /Add/ }));

        await waitFor(() => expect(console.error).toHaveBeenCalled());
    });

    it("clears the opening flag once the second factor dialog reports it opened", async () => {
        renderPanel();

        fireEvent.click(screen.getByRole("button", { name: /Add/ }));
        await waitFor(() => expect(screen.getByTestId("second-factor-dialog")).toHaveAttribute("data-opening", "true"));

        fireEvent.click(screen.getByTestId("sf-opened"));

        await waitFor(() =>
            expect(screen.getByTestId("second-factor-dialog")).toHaveAttribute("data-opening", "false"),
        );
    });

    it("clears the opening flag once the identity dialog reports it opened", async () => {
        getElevationMock.mockResolvedValue(notElevated);

        renderPanel();

        fireEvent.click(screen.getByRole("button", { name: /Add/ }));
        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());

        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));
        await waitFor(() => expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "true"));

        fireEvent.click(screen.getByTestId("iv-opened"));

        await waitFor(() => expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "false"));
    });
});

describe("second factor dialog result", () => {
    it("resets the state when the user cancels", async () => {
        renderPanel();

        fireEvent.click(screen.getByRole("button", { name: /Add/ }));
        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());

        fireEvent.click(screen.getByTestId("sf-closed-cancel"));

        await waitFor(() => expect(console.warn).toHaveBeenCalled());
        expect(screen.getByTestId("register-dialog")).toHaveAttribute("data-open", "false");
        await waitFor(() => expect(screen.getByRole("button", { name: /Add/ })).not.toBeDisabled());
    });

    it("opens the register dialog when already elevated and nothing changed", async () => {
        renderPanel();

        fireEvent.click(screen.getByRole("button", { name: /Add/ }));
        await waitFor(() =>
            expect(screen.getByTestId("second-factor-dialog")).toHaveAttribute(
                "data-elevation",
                JSON.stringify(elevated),
            ),
        );

        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));

        await waitFor(() => expect(screen.getByTestId("register-dialog")).toHaveAttribute("data-open", "true"));
    });

    it("treats skip_second_factor as elevated", async () => {
        getElevationMock.mockResolvedValue(skipSecondFactor);

        renderPanel();

        fireEvent.click(screen.getByRole("button", { name: /Add/ }));
        await waitFor(() =>
            expect(screen.getByTestId("second-factor-dialog")).toHaveAttribute(
                "data-elevation",
                JSON.stringify(skipSecondFactor),
            ),
        );

        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));

        await waitFor(() => expect(screen.getByTestId("register-dialog")).toHaveAttribute("data-open", "true"));
    });

    it("falls through to identity verification when not elevated", async () => {
        getElevationMock.mockResolvedValue(notElevated);

        renderPanel();

        fireEvent.click(screen.getByRole("button", { name: /Add/ }));
        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());

        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));

        await waitFor(() => expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "true"));
        expect(screen.getByTestId("register-dialog")).toHaveAttribute("data-open", "false");
    });

    it("re-checks elevation when the dialog reports a change", async () => {
        getElevationMock.mockResolvedValueOnce(notElevated).mockResolvedValueOnce(elevated);

        renderPanel();

        fireEvent.click(screen.getByRole("button", { name: /Add/ }));
        await waitFor(() => expect(getElevationMock).toHaveBeenCalledTimes(1));

        fireEvent.click(screen.getByTestId("sf-closed-ok-changed"));

        await waitFor(() => expect(getElevationMock).toHaveBeenCalledTimes(2));
        await waitFor(() => expect(screen.getByTestId("register-dialog")).toHaveAttribute("data-open", "true"));
    });

    it("asks for identity verification when the refreshed elevation is insufficient", async () => {
        getElevationMock.mockResolvedValue(notElevated);

        renderPanel();

        fireEvent.click(screen.getByRole("button", { name: /Add/ }));
        await waitFor(() => expect(getElevationMock).toHaveBeenCalledTimes(1));

        fireEvent.click(screen.getByTestId("sf-closed-ok-changed"));

        await waitFor(() => expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "true"));
    });

    it("does nothing when the elevation refresh fails after a change", async () => {
        getElevationMock.mockResolvedValueOnce(elevated).mockRejectedValueOnce(new Error("boom"));

        renderPanel();

        fireEvent.click(screen.getByRole("button", { name: /Add/ }));
        await waitFor(() => expect(getElevationMock).toHaveBeenCalledTimes(1));

        fireEvent.click(screen.getByTestId("sf-closed-ok-changed"));

        await waitFor(() => expect(console.error).toHaveBeenCalled());
        expect(screen.getByTestId("register-dialog")).toHaveAttribute("data-open", "false");
    });

    it("opens the edit dialog for the selected credential", async () => {
        renderPanel();

        fireEvent.click(screen.getByTestId("grid-edit"));
        await waitFor(() =>
            expect(screen.getByTestId("second-factor-dialog")).toHaveAttribute(
                "data-elevation",
                JSON.stringify(elevated),
            ),
        );

        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));

        await waitFor(() => expect(screen.getByTestId("edit-dialog")).toHaveAttribute("data-open", "true"));
        expect(screen.getByTestId("edit-dialog")).toHaveAttribute("data-credential", JSON.stringify(credentials[0]));
    });

    it("opens the delete dialog for the selected credential", async () => {
        renderPanel();

        fireEvent.click(screen.getByTestId("grid-delete"));
        await waitFor(() =>
            expect(screen.getByTestId("second-factor-dialog")).toHaveAttribute(
                "data-elevation",
                JSON.stringify(elevated),
            ),
        );

        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));

        await waitFor(() => expect(screen.getByTestId("delete-dialog")).toHaveAttribute("data-open", "true"));
        expect(screen.getByTestId("delete-dialog")).toHaveAttribute("data-credential", JSON.stringify(credentials[0]));
    });
});

describe("identity verification dialog result", () => {
    it("resets the state when the user cancels", async () => {
        getElevationMock.mockResolvedValue(notElevated);

        renderPanel();

        fireEvent.click(screen.getByRole("button", { name: /Add/ }));
        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());
        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));
        await waitFor(() => expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "true"));

        fireEvent.click(screen.getByTestId("iv-closed-cancel"));

        await waitFor(() => expect(screen.getByTestId("register-dialog")).toHaveAttribute("data-open", "false"));
        expect(console.warn).toHaveBeenCalled();
    });

    it("opens the register dialog once identity is verified", async () => {
        getElevationMock.mockResolvedValue(notElevated);

        renderPanel();

        fireEvent.click(screen.getByRole("button", { name: /Add/ }));
        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());
        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));
        await waitFor(() => expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "true"));

        fireEvent.click(screen.getByTestId("iv-closed-ok"));

        await waitFor(() => expect(screen.getByTestId("register-dialog")).toHaveAttribute("data-open", "true"));
    });

    it("opens the edit dialog once identity is verified", async () => {
        getElevationMock.mockResolvedValue(notElevated);

        renderPanel();

        fireEvent.click(screen.getByTestId("grid-edit"));
        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());
        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));
        await waitFor(() => expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "true"));

        fireEvent.click(screen.getByTestId("iv-closed-ok"));

        await waitFor(() => expect(screen.getByTestId("edit-dialog")).toHaveAttribute("data-open", "true"));
    });

    it("opens the delete dialog once identity is verified", async () => {
        getElevationMock.mockResolvedValue(notElevated);

        renderPanel();

        fireEvent.click(screen.getByTestId("grid-delete"));
        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());
        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));
        await waitFor(() => expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "true"));

        fireEvent.click(screen.getByTestId("iv-closed-ok"));

        await waitFor(() => expect(screen.getByTestId("delete-dialog")).toHaveAttribute("data-open", "true"));
    });
});

describe("information dialog", () => {
    it("opens with the selected credential without requiring elevation", async () => {
        renderPanel();

        fireEvent.click(screen.getByTestId("grid-information"));

        await waitFor(() => expect(screen.getByTestId("info-dialog")).toHaveAttribute("data-open", "true"));
        expect(screen.getByTestId("info-dialog")).toHaveAttribute("data-credential", JSON.stringify(credentials[0]));
        expect(getElevationMock).not.toHaveBeenCalled();
    });

    it("closes without refreshing the parent state", async () => {
        const { handleRefreshState } = renderPanel();

        fireEvent.click(screen.getByTestId("grid-information"));
        await waitFor(() => expect(screen.getByTestId("info-dialog")).toHaveAttribute("data-open", "true"));

        fireEvent.click(screen.getByTestId("info-close"));

        await waitFor(() => expect(screen.getByTestId("info-dialog")).toHaveAttribute("data-open", "false"));
        expect(handleRefreshState).not.toHaveBeenCalled();
    });

    it("ignores an out of range index", () => {
        renderPanel();

        fireEvent.click(screen.getByTestId("grid-information-oob"));

        expect(screen.getByTestId("info-dialog")).toHaveAttribute("data-open", "false");
    });
});

describe("guards against missing credentials", () => {
    it("does not render the credentials grid when credentials are not loaded", () => {
        renderPanel({ credentials: undefined });

        expect(screen.queryByTestId("credentials-grid")).not.toBeInTheDocument();
        expect(getElevationMock).not.toHaveBeenCalled();
    });

    it("ignores an out of range edit index", () => {
        renderPanel();

        fireEvent.click(screen.getByTestId("grid-edit-oob"));

        expect(getElevationMock).not.toHaveBeenCalled();
    });

    it("ignores an out of range delete index", () => {
        renderPanel();

        fireEvent.click(screen.getByTestId("grid-delete-oob"));

        expect(getElevationMock).not.toHaveBeenCalled();
    });
});

describe("refreshing after changes", () => {
    it("refreshes the parent state when the register dialog closes", async () => {
        const { handleRefreshState } = renderPanel();

        fireEvent.click(screen.getByTestId("register-close"));

        await waitFor(() => expect(handleRefreshState).toHaveBeenCalled());
    });

    it("refreshes the parent state when the edit dialog closes", async () => {
        const { handleRefreshState } = renderPanel();

        fireEvent.click(screen.getByTestId("edit-close"));

        await waitFor(() => expect(handleRefreshState).toHaveBeenCalled());
    });

    it("refreshes the parent state when the delete dialog closes", async () => {
        const { handleRefreshState } = renderPanel();

        fireEvent.click(screen.getByTestId("delete-close"));

        await waitFor(() => expect(handleRefreshState).toHaveBeenCalled());
    });
});

describe("elevation refresh after a change", () => {
    it("opens the edit dialog when the refreshed elevation is sufficient", async () => {
        getElevationMock.mockResolvedValueOnce(notElevated).mockResolvedValueOnce(elevated);

        renderPanel();

        fireEvent.click(screen.getByTestId("grid-edit"));
        await waitFor(() => expect(getElevationMock).toHaveBeenCalledTimes(1));

        fireEvent.click(screen.getByTestId("sf-closed-ok-changed"));

        await waitFor(() => expect(screen.getByTestId("edit-dialog")).toHaveAttribute("data-open", "true"));
    });

    it("opens the delete dialog when the refreshed elevation is sufficient", async () => {
        getElevationMock.mockResolvedValueOnce(notElevated).mockResolvedValueOnce(elevated);

        renderPanel();

        fireEvent.click(screen.getByTestId("grid-delete"));
        await waitFor(() => expect(getElevationMock).toHaveBeenCalledTimes(1));

        fireEvent.click(screen.getByTestId("sf-closed-ok-changed"));

        await waitFor(() => expect(screen.getByTestId("delete-dialog")).toHaveAttribute("data-open", "true"));
    });

    it("opens no dialog when nothing was pending", async () => {
        getElevationMock.mockResolvedValue(elevated);

        renderPanel();

        fireEvent.click(screen.getByTestId("sf-closed-ok-changed"));

        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());
        expect(screen.getByTestId("register-dialog")).toHaveAttribute("data-open", "false");
        expect(screen.getByTestId("edit-dialog")).toHaveAttribute("data-open", "false");
        expect(screen.getByTestId("delete-dialog")).toHaveAttribute("data-open", "false");
    });

    it("opens no dialog when identity verification succeeds without a pending action", async () => {
        renderPanel();

        fireEvent.click(screen.getByTestId("iv-closed-ok"));

        await waitFor(() => expect(screen.getByTestId("register-dialog")).toHaveAttribute("data-open", "false"));
        expect(screen.getByTestId("edit-dialog")).toHaveAttribute("data-open", "false");
        expect(screen.getByTestId("delete-dialog")).toHaveAttribute("data-open", "false");
    });

    it("passes no credential to the dialogs while credentials are unavailable", () => {
        renderPanel({ credentials: undefined });

        expect(screen.getByTestId("edit-dialog")).toHaveAttribute("data-credential", "null");
        expect(screen.getByTestId("delete-dialog")).toHaveAttribute("data-credential", "null");
        expect(screen.getByTestId("info-dialog")).toHaveAttribute("data-credential", "null");
    });
});
