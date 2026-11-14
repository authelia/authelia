// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import RecoveryCodesGenerateDialog from "@views/Settings/TwoFactorAuthentication/RecoveryCodesGenerateDialog";

const mocks = vi.hoisted(() => ({
    error: undefined as Error | undefined,
    inProgress: false,
    response: undefined as any,
    trigger: vi.fn(),
}));

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@hooks/UserRecoveryCodes", () => ({
    useGenerateRecoveryCodes: () => [mocks.response, mocks.trigger, mocks.inProgress, mocks.error],
}));

vi.mock("@components/CopyButton", () => ({
    default: (props: any) => <button data-testid="copy-button" data-value={props.value} />,
}));

const codes = ["AAAA-BBBB-CCCC", "DDDD-EEEE-FFFF"];

const writeTextMock = vi.fn();

function renderDialog(open = true) {
    const setClosed = vi.fn();
    const result = render(<RecoveryCodesGenerateDialog open={open} setClosed={setClosed} />);

    return {
        ...result,
        rerenderDialog: (nextOpen = open) =>
            result.rerender(<RecoveryCodesGenerateDialog open={nextOpen} setClosed={setClosed} />),
        setClosed,
    };
}

function getDone() {
    return document.getElementById("recovery-codes-done") as HTMLButtonElement;
}

function getConfirmInput() {
    return document.getElementById("recovery-codes-confirm-input") as HTMLInputElement;
}

function pressEscape() {
    fireEvent.keyDown(document.activeElement ?? document.body, { key: "Escape" });
}

beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, "error").mockImplementation(() => {});
    mocks.error = undefined;
    mocks.inProgress = false;
    mocks.response = { codes, notification_sent: true };
    writeTextMock.mockResolvedValue(undefined);
    Object.defineProperty(navigator, "clipboard", { configurable: true, value: { writeText: writeTextMock } });
});

afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
});

describe("generation", () => {
    it("triggers generation once when opened", () => {
        const { rerenderDialog } = renderDialog();

        rerenderDialog();

        expect(mocks.trigger).toHaveBeenCalledTimes(1);
    });

    it("does not trigger generation while closed", () => {
        renderDialog(false);

        expect(mocks.trigger).not.toHaveBeenCalled();
    });

    it("triggers generation when it later opens", () => {
        const { rerenderDialog } = renderDialog(false);

        rerenderDialog(true);

        expect(mocks.trigger).toHaveBeenCalledTimes(1);
    });
});

describe("rendering", () => {
    it("shows a spinner and disables Done while loading", () => {
        mocks.inProgress = true;
        mocks.response = undefined;

        renderDialog();

        expect(screen.getByText("Save your recovery codes")).toBeInTheDocument();
        expect(document.querySelector('[data-slot="spinner"]')).toBeInTheDocument();
        expect(getDone()).toBeDisabled();
        expect(getConfirmInput()).toBeNull();
    });

    it("renders each generated code with a copy button", () => {
        renderDialog();

        expect(screen.getByText("AAAA-BBBB-CCCC")).toBeInTheDocument();
        expect(screen.getByText("DDDD-EEEE-FFFF")).toBeInTheDocument();
        expect(screen.getAllByTestId("copy-button").map((el) => el.getAttribute("data-value"))).toEqual(codes);
        expect(document.querySelector('[data-slot="spinner"]')).toBeNull();
    });

    it("does not warn when the notification was sent", () => {
        renderDialog();

        expect(
            screen.queryByText(
                "We could not email you a confirmation; if you did not initiate this, change your password immediately.",
            ),
        ).not.toBeInTheDocument();
    });

    it("warns when the notification could not be sent", () => {
        mocks.response = { codes, notification_sent: false };

        renderDialog();

        expect(
            screen.getByText(
                "We could not email you a confirmation; if you did not initiate this, change your password immediately.",
            ),
        ).toBeInTheDocument();
    });

    it("shows the error and enables Done when generation fails", () => {
        mocks.error = new Error("boom");
        mocks.response = undefined;

        renderDialog();

        expect(screen.getByText("Failed to generate recovery codes")).toBeInTheDocument();
        expect(getDone()).not.toBeDisabled();
        expect(getConfirmInput()).toBeNull();
    });

    it("keeps Done disabled while loading even after an error", () => {
        mocks.error = new Error("boom");
        mocks.inProgress = true;
        mocks.response = undefined;

        renderDialog();

        expect(getDone()).toBeDisabled();
    });
});

describe("confirmation", () => {
    it("keeps Done disabled until a code is confirmed", () => {
        renderDialog();

        expect(getDone()).toBeDisabled();

        fireEvent.change(getConfirmInput(), { target: { value: "AAAA-BBBB-XXXX" } });

        expect(getDone()).toBeDisabled();
    });

    it("enables Done once a code is typed exactly", () => {
        renderDialog();

        fireEvent.change(getConfirmInput(), { target: { value: "DDDD-EEEE-FFFF" } });

        expect(getDone()).not.toBeDisabled();
    });

    it("normalizes case, whitespace, dashes and underscores when confirming", () => {
        renderDialog();

        fireEvent.change(getConfirmInput(), { target: { value: " aaaa_bbbb cccc " } });

        expect(getDone()).not.toBeDisabled();
    });

    it("closes and clears the confirmation after confirming", () => {
        const { setClosed } = renderDialog();

        fireEvent.change(getConfirmInput(), { target: { value: "AAAA-BBBB-CCCC" } });
        fireEvent.click(getDone());

        expect(setClosed).toHaveBeenCalledTimes(1);
        expect(getConfirmInput()).toHaveValue("");
        expect(getDone()).toBeDisabled();
    });

    it("closes after an error", () => {
        mocks.error = new Error("boom");
        mocks.response = undefined;

        const { setClosed } = renderDialog();

        fireEvent.click(getDone());

        expect(setClosed).toHaveBeenCalledTimes(1);
    });
});

describe("dismissal", () => {
    it("ignores Escape before a code is confirmed", async () => {
        const { setClosed } = renderDialog();
        await screen.findByText("AAAA-BBBB-CCCC");

        pressEscape();

        expect(setClosed).not.toHaveBeenCalled();
        expect(screen.getByText("AAAA-BBBB-CCCC")).toBeInTheDocument();
    });

    it("ignores Escape while loading", async () => {
        mocks.inProgress = true;
        mocks.response = undefined;
        mocks.error = new Error("boom");

        const { setClosed } = renderDialog();
        await screen.findByText("Save your recovery codes");

        pressEscape();

        expect(setClosed).not.toHaveBeenCalled();
    });

    it("closes on Escape after a code is confirmed", async () => {
        const { setClosed } = renderDialog();

        fireEvent.change(getConfirmInput(), { target: { value: "AAAA-BBBB-CCCC" } });
        pressEscape();

        await waitFor(() => expect(setClosed).toHaveBeenCalledTimes(1));
    });

    it("closes on Escape after an error", async () => {
        mocks.error = new Error("boom");
        mocks.response = undefined;

        const { setClosed } = renderDialog();
        await screen.findByText("Failed to generate recovery codes");

        pressEscape();

        await waitFor(() => expect(setClosed).toHaveBeenCalledTimes(1));
    });
});

describe("actions", () => {
    it("copies all codes separated by newlines", () => {
        renderDialog();

        fireEvent.click(document.getElementById("recovery-codes-copy-all") as HTMLButtonElement);

        expect(writeTextMock).toHaveBeenCalledWith("AAAA-BBBB-CCCC\nDDDD-EEEE-FFFF");
    });

    it("logs clipboard failures", async () => {
        writeTextMock.mockRejectedValue(new Error("denied"));

        renderDialog();

        fireEvent.click(document.getElementById("recovery-codes-copy-all") as HTMLButtonElement);

        await waitFor(() => expect(console.error).toHaveBeenCalledWith(new Error("denied")));
    });

    it("downloads the codes as a dated text file", async () => {
        vi.useFakeTimers({ toFake: ["Date"] });
        vi.setSystemTime(new Date(2026, 2, 4, 5, 6, 7));

        const createObjectURL = vi.fn((_blob: Blob) => "blob:recovery-codes");
        const revokeObjectURL = vi.fn();
        Object.defineProperty(URL, "createObjectURL", { configurable: true, value: createObjectURL });
        Object.defineProperty(URL, "revokeObjectURL", { configurable: true, value: revokeObjectURL });

        let clicked: HTMLAnchorElement | undefined;
        vi.spyOn(HTMLAnchorElement.prototype, "click").mockImplementation(function (this: HTMLAnchorElement) {
            clicked = this;
        });

        renderDialog();

        fireEvent.click(document.getElementById("recovery-codes-download") as HTMLButtonElement);

        expect(clicked?.download).toBe("authelia-recovery-codes-2026-03-04.txt");
        expect(clicked?.href).toBe("blob:recovery-codes");
        expect(clicked?.isConnected).toBe(false);
        expect(revokeObjectURL).toHaveBeenCalledWith("blob:recovery-codes");

        const blob = createObjectURL.mock.calls[0][0];
        expect(blob.type).toBe("text/plain");
        expect(await blob.text()).toBe(
            [
                "Authelia recovery codes",
                `Generated: ${new Date(2026, 2, 4, 5, 6, 7).toISOString()}`,
                "Anyone with this file can sign in to your account. Treat it like a password.",
                "AAAA-BBBB-CCCC",
                "DDDD-EEEE-FFFF",
                "",
            ].join("\n"),
        );
    });
});
