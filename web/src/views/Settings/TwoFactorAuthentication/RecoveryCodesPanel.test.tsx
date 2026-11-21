// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import { getUserSessionElevation } from "@services/UserSessionElevation";
import RecoveryCodesPanel from "@views/Settings/TwoFactorAuthentication/RecoveryCodesPanel";

const mocks = vi.hoisted(() => ({
    createErrorNotification: vi.fn(),
    error: undefined as Error | undefined,
    fetchStatus: vi.fn(),
    loading: false,
    status: undefined as any,
    translate: (key: string, options?: Record<string, unknown>) =>
        key.replace(/{{(\w+)}}/g, (_match: string, name: string) => String(options?.[name])),
}));

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: mocks.translate }),
}));

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({ createErrorNotification: mocks.createErrorNotification }),
}));

vi.mock("@hooks/UserRecoveryCodes", () => ({
    useRecoveryCodesStatus: () => [mocks.status, mocks.fetchStatus, mocks.loading, mocks.error],
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

vi.mock("@views/Settings/TwoFactorAuthentication/RecoveryCodesGenerateDialog", () => ({
    default: (props: any) => (
        <div data-testid="generate-dialog" data-open={String(props.open)}>
            <button data-testid="generate-close" onClick={() => props.setClosed()} />
        </div>
    ),
}));

const getElevationMock = vi.mocked(getUserSessionElevation);

const elevated = { elevated: true, skip_second_factor: false } as any;
const notElevated = { elevated: false, skip_second_factor: false } as any;
const skipSecondFactor = { elevated: false, skip_second_factor: true } as any;

const lowWarning = "Recovery codes are running low. Regenerate to get a fresh batch before you run out.";
const noCodes = "No recovery codes have been generated yet.";

function status(overrides: Record<string, unknown> = {}) {
    return { codes_remaining: 10, codes_total: 10, generated_at: undefined, last_used_at: undefined, ...overrides };
}

function renderPanel() {
    return render(<RecoveryCodesPanel info={undefined} />);
}

function getGenerate() {
    return document.getElementById("recovery-codes-add") as HTMLButtonElement;
}

async function openSecondFactor() {
    fireEvent.click(getGenerate());
    await waitFor(() => expect(getElevationMock).toHaveBeenCalled());
    await waitFor(() => expect(screen.getByTestId("second-factor-dialog")).toHaveAttribute("data-opening", "true"));
}

beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, "error").mockImplementation(() => {});
    mocks.error = undefined;
    mocks.loading = false;
    mocks.status = undefined;
    getElevationMock.mockResolvedValue(elevated);
});

afterEach(() => {
    vi.restoreAllMocks();
});

describe("status", () => {
    it("fetches the status on mount", () => {
        renderPanel();

        expect(mocks.fetchStatus).toHaveBeenCalledTimes(1);
        expect(screen.getByText("Recovery Codes")).toBeInTheDocument();
    });

    it("marks the panel as loading and disables the button", () => {
        mocks.loading = true;

        renderPanel();

        expect(document.getElementById("recovery-codes-panel")).toHaveAttribute("data-loading", "true");
        expect(getGenerate()).toBeDisabled();
    });

    it("offers generation when no codes exist", () => {
        mocks.status = status({ codes_remaining: 0, codes_total: 0 });

        renderPanel();

        expect(document.getElementById("recovery-codes-panel")).toHaveAttribute("data-loading", "false");
        expect(screen.getByText(noCodes)).toBeInTheDocument();
        expect(getGenerate()).toHaveTextContent("Generate");
        expect(getGenerate()).not.toHaveTextContent("Regenerate");
        expect(getGenerate()).not.toBeDisabled();
        expect(screen.queryByText(lowWarning)).not.toBeInTheDocument();
    });

    it("offers generation before the status is known", () => {
        renderPanel();

        expect(screen.getByText(noCodes)).toBeInTheDocument();
        expect(getGenerate()).toHaveTextContent("Generate");
    });

    it("shows usage counts and timestamps", () => {
        const generatedAt = new Date("2026-01-02T03:04:05Z");
        const lastUsedAt = new Date("2026-02-03T04:05:06Z");
        mocks.status = status({ codes_remaining: 7, generated_at: generatedAt, last_used_at: lastUsedAt });

        renderPanel();

        expect(screen.getByText("3 of 10 codes used")).toBeInTheDocument();
        expect(screen.getByText(`Generated ${generatedAt.toLocaleString()}`)).toBeInTheDocument();
        expect(screen.getByText(`Last used ${lastUsedAt.toLocaleString()}`)).toBeInTheDocument();
        expect(getGenerate()).toHaveTextContent("Regenerate");
        expect(screen.queryByText(noCodes)).not.toBeInTheDocument();
    });

    it("shows that unused codes have never been used", () => {
        mocks.status = status();

        renderPanel();

        expect(screen.getByText("0 of 10 codes used")).toBeInTheDocument();
        expect(screen.getByText("Never used")).toBeInTheDocument();
        expect(screen.queryByText(/^Generated /)).not.toBeInTheDocument();
    });

    it.each([1, 2])("warns when %i codes remain", (remaining) => {
        mocks.status = status({ codes_remaining: remaining });

        renderPanel();

        expect(screen.getByText(lowWarning)).toBeInTheDocument();
    });

    it("does not warn when more than two codes remain", () => {
        mocks.status = status({ codes_remaining: 3 });

        renderPanel();

        expect(screen.queryByText(lowWarning)).not.toBeInTheDocument();
    });

    it("does not show the low warning once every code is used", () => {
        mocks.status = status({ codes_remaining: 0 });

        renderPanel();

        expect(screen.getByText("10 of 10 codes used")).toBeInTheDocument();
        expect(screen.queryByText(lowWarning)).not.toBeInTheDocument();
        expect(getGenerate()).toHaveTextContent("Regenerate");
    });

    it("notifies and hides the controls when the status cannot be retrieved", () => {
        mocks.error = new Error("boom");

        renderPanel();

        expect(mocks.createErrorNotification).toHaveBeenCalledWith("There was an issue retrieving the Recovery Codes");
        expect(getGenerate()).toBeNull();
        expect(screen.queryByText(noCodes)).not.toBeInTheDocument();
    });

    it("keeps showing a previously retrieved status after a retrieval error", () => {
        mocks.error = new Error("boom");
        mocks.status = status({ codes_remaining: 4 });

        renderPanel();

        expect(mocks.createErrorNotification).toHaveBeenCalled();
        expect(screen.getByText("6 of 10 codes used")).toBeInTheDocument();
        expect(getGenerate()).not.toBeNull();
    });
});

describe("regeneration confirmation", () => {
    it("starts the flow immediately when no codes exist", async () => {
        renderPanel();

        await openSecondFactor();

        expect(screen.queryByText("Regenerate recovery codes?")).not.toBeInTheDocument();
        expect(getGenerate()).toBeDisabled();
        expect(getGenerate().querySelector('[data-slot="spinner"]')).toBeInTheDocument();
    });

    it("asks for confirmation before regenerating existing codes", async () => {
        mocks.status = status();

        renderPanel();

        fireEvent.click(getGenerate());

        expect(await screen.findByText("Regenerate recovery codes?")).toBeInTheDocument();
        expect(getElevationMock).not.toHaveBeenCalled();
        expect(screen.getByTestId("second-factor-dialog")).toHaveAttribute("data-opening", "false");
    });

    it("does nothing when the confirmation is cancelled", async () => {
        mocks.status = status();

        renderPanel();

        fireEvent.click(getGenerate());
        await screen.findByText("Regenerate recovery codes?");

        fireEvent.click(document.getElementById("recovery-codes-confirm-cancel") as HTMLButtonElement);

        await waitFor(() => expect(screen.queryByText("Regenerate recovery codes?")).not.toBeInTheDocument());
        expect(getElevationMock).not.toHaveBeenCalled();
    });

    it("closes the confirmation on Escape", async () => {
        mocks.status = status();

        renderPanel();

        fireEvent.click(getGenerate());
        await screen.findByText("Regenerate recovery codes?");

        fireEvent.keyDown(document.activeElement ?? document.body, { key: "Escape" });

        await waitFor(() => expect(screen.queryByText("Regenerate recovery codes?")).not.toBeInTheDocument());
        expect(getElevationMock).not.toHaveBeenCalled();
    });

    it("starts the flow once regeneration is confirmed", async () => {
        mocks.status = status();

        renderPanel();

        fireEvent.click(getGenerate());
        await screen.findByText("Regenerate recovery codes?");

        fireEvent.click(document.getElementById("recovery-codes-confirm-accept") as HTMLButtonElement);

        await waitFor(() => expect(getElevationMock).toHaveBeenCalled());
        expect(screen.getByTestId("second-factor-dialog")).toHaveAttribute("data-opening", "true");
        await waitFor(() => expect(screen.queryByText("Regenerate recovery codes?")).not.toBeInTheDocument());
    });
});

describe("elevation flow", () => {
    it("opens the generate dialog when already elevated", async () => {
        renderPanel();
        await openSecondFactor();
        await waitFor(() =>
            expect(screen.getByTestId("second-factor-dialog")).toHaveAttribute(
                "data-elevation",
                JSON.stringify(elevated),
            ),
        );

        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));

        expect(await screen.findByTestId("generate-dialog")).toHaveAttribute("data-open", "true");
        expect(getGenerate()).toBeDisabled();
    });

    it("treats skip_second_factor as elevated", async () => {
        getElevationMock.mockResolvedValue(skipSecondFactor);

        renderPanel();
        await openSecondFactor();
        await waitFor(() =>
            expect(screen.getByTestId("second-factor-dialog")).toHaveAttribute(
                "data-elevation",
                JSON.stringify(skipSecondFactor),
            ),
        );

        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));

        expect(await screen.findByTestId("generate-dialog")).toBeInTheDocument();
    });

    it("asks for identity verification when not elevated", async () => {
        getElevationMock.mockResolvedValue(notElevated);

        renderPanel();
        await openSecondFactor();

        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));

        await waitFor(() => expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "true"));
        expect(screen.queryByTestId("generate-dialog")).not.toBeInTheDocument();
    });

    it("asks for identity verification when the elevation lookup failed", async () => {
        getElevationMock.mockRejectedValue(new Error("boom"));

        renderPanel();
        await openSecondFactor();
        await waitFor(() => expect(console.error).toHaveBeenCalled());

        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));

        await waitFor(() => expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "true"));
    });

    it("re-checks elevation when the second factor dialog reports a change", async () => {
        getElevationMock.mockResolvedValueOnce(notElevated).mockResolvedValueOnce(elevated);

        renderPanel();
        await openSecondFactor();

        fireEvent.click(screen.getByTestId("sf-closed-ok-changed"));

        await waitFor(() => expect(getElevationMock).toHaveBeenCalledTimes(2));
        expect(await screen.findByTestId("generate-dialog")).toBeInTheDocument();
    });

    it("asks for identity verification when the refreshed elevation is insufficient", async () => {
        getElevationMock.mockResolvedValue(notElevated);

        renderPanel();
        await openSecondFactor();

        fireEvent.click(screen.getByTestId("sf-closed-ok-changed"));

        await waitFor(() => expect(getElevationMock).toHaveBeenCalledTimes(2));
        await waitFor(() => expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "true"));
    });

    it("does nothing when the elevation refresh fails after a change", async () => {
        getElevationMock.mockResolvedValueOnce(elevated).mockRejectedValueOnce(new Error("boom"));

        renderPanel();
        await openSecondFactor();

        fireEvent.click(screen.getByTestId("sf-closed-ok-changed"));

        await waitFor(() => expect(console.error).toHaveBeenCalled());
        expect(screen.queryByTestId("generate-dialog")).not.toBeInTheDocument();
        expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "false");
    });

    it("resets when the second factor dialog is cancelled", async () => {
        renderPanel();
        await openSecondFactor();

        fireEvent.click(screen.getByTestId("sf-closed-cancel"));

        await waitFor(() => expect(getGenerate()).not.toBeDisabled());
        expect(screen.getByTestId("second-factor-dialog")).toHaveAttribute("data-opening", "false");
        expect(screen.getByTestId("second-factor-dialog")).toHaveAttribute("data-elevation", "null");
        expect(screen.queryByTestId("generate-dialog")).not.toBeInTheDocument();
    });

    it("clears the second factor opening flag once it opens", async () => {
        renderPanel();
        await openSecondFactor();

        fireEvent.click(screen.getByTestId("sf-opened"));

        await waitFor(() =>
            expect(screen.getByTestId("second-factor-dialog")).toHaveAttribute("data-opening", "false"),
        );
    });

    it("opens the generate dialog once identity is verified", async () => {
        getElevationMock.mockResolvedValue(notElevated);

        renderPanel();
        await openSecondFactor();
        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));
        await waitFor(() => expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "true"));

        fireEvent.click(screen.getByTestId("iv-closed-ok"));

        expect(await screen.findByTestId("generate-dialog")).toBeInTheDocument();
    });

    it("resets when identity verification is cancelled", async () => {
        getElevationMock.mockResolvedValue(notElevated);

        renderPanel();
        await openSecondFactor();
        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));
        await waitFor(() => expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "true"));

        fireEvent.click(screen.getByTestId("iv-closed-cancel"));

        await waitFor(() => expect(getGenerate()).not.toBeDisabled());
        expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "false");
        expect(screen.queryByTestId("generate-dialog")).not.toBeInTheDocument();
    });

    it("clears the identity verification opening flag once it opens", async () => {
        getElevationMock.mockResolvedValue(notElevated);

        renderPanel();
        await openSecondFactor();
        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));
        await waitFor(() => expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "true"));

        fireEvent.click(screen.getByTestId("iv-opened"));

        await waitFor(() => expect(screen.getByTestId("identity-dialog")).toHaveAttribute("data-opening", "false"));
    });

    it("refreshes the status when the generate dialog closes", async () => {
        renderPanel();
        await openSecondFactor();
        fireEvent.click(screen.getByTestId("sf-closed-ok-unchanged"));
        await screen.findByTestId("generate-dialog");
        expect(mocks.fetchStatus).toHaveBeenCalledTimes(1);

        fireEvent.click(screen.getByTestId("generate-close"));

        await waitFor(() => expect(screen.queryByTestId("generate-dialog")).not.toBeInTheDocument());
        expect(mocks.fetchStatus).toHaveBeenCalledTimes(2);
        expect(getGenerate()).not.toBeDisabled();
        expect(getGenerate().querySelector('[data-slot="spinner"]')).toBeNull();
    });
});
