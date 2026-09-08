import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";

import {
    deleteUserSessionElevation,
    generateUserSessionElevation,
    verifyUserSessionElevation,
} from "@services/UserSessionElevation";
import IdentityVerificationDialog from "@views/Settings/Common/IdentityVerificationDialog";

const mocks = vi.hoisted(() => ({
    createErrorNotification: vi.fn(),
}));

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: mocks.createErrorNotification,
    }),
}));

vi.mock("@components/OneTimeCodeTextField", () => ({
    default: (props: any) => (
        <input
            data-testid="one-time-code"
            id={props.id}
            ref={props.inputRef}
            value={props.value}
            aria-invalid={props.error}
            disabled={props.disabled}
            onChange={props.onChange}
            onKeyDown={props.onKeyDown}
        />
    ),
}));

vi.mock("@components/SuccessIcon", () => ({
    default: () => <div data-testid="success-icon" />,
}));

vi.mock("@services/UserSessionElevation", () => ({
    deleteUserSessionElevation: vi.fn(),
    generateUserSessionElevation: vi.fn(),
    verifyUserSessionElevation: vi.fn(),
}));

const generateMock = vi.mocked(generateUserSessionElevation);
const verifyMock = vi.mocked(verifyUserSessionElevation);
const deleteMock = vi.mocked(deleteUserSessionElevation);

const elevation = { elevated: false, skip_second_factor: false } as any;

function renderDialog(
    props: Partial<{
        elevation: any;
        handleClosed: (ok: boolean) => void;
        handleOpened: () => void;
        opening: boolean;
    }> = {},
) {
    const handleClosed = props.handleClosed ?? vi.fn();
    const handleOpened = props.handleOpened ?? vi.fn();
    const merged = { elevation, opening: true, ...props, handleClosed, handleOpened };

    return { ...render(<IdentityVerificationDialog {...merged} />), handleClosed, handleOpened };
}

function getCode() {
    return document.getElementById("one-time-code") as HTMLInputElement;
}

function clickVerify() {
    fireEvent.click(document.getElementById("dialog-verify") as HTMLButtonElement);
}

function clickCancel() {
    fireEvent.click(document.getElementById("dialog-cancel") as HTMLButtonElement);
}

beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, "error").mockImplementation(() => {});
    generateMock.mockResolvedValue({ delete_id: "delete-id" } as any);
    verifyMock.mockResolvedValue(true as any);
    deleteMock.mockResolvedValue(undefined as any);
});

afterEach(() => {
    vi.restoreAllMocks();
    vi.useRealTimers();
});

describe("rendering", () => {
    it("renders dialog with title when opening and elevation resolve", async () => {
        renderDialog();
        expect(await screen.findByText("Identity Verification")).toBeInTheDocument();
    });

    it("renders cancel and verify buttons after elevation generation resolves", async () => {
        renderDialog();
        expect(await screen.findByText("Cancel")).toBeInTheDocument();
        expect(screen.getByText("Verify")).toBeInTheDocument();
    });

    it("does not render content when not opening", () => {
        renderDialog({ opening: false });
        expect(screen.queryByText("Identity Verification")).not.toBeInTheDocument();
        expect(generateMock).not.toHaveBeenCalled();
    });

    it("does not generate a code without an elevation", () => {
        renderDialog({ elevation: undefined });
        expect(generateMock).not.toHaveBeenCalled();
    });

    it("notifies the parent that it opened", async () => {
        const { handleOpened } = renderDialog();
        await waitFor(() => expect(handleOpened).toHaveBeenCalled());
    });

    it("explains that closing invalidates the code", async () => {
        renderDialog();
        expect(
            await screen.findByText("Closing this dialog or selecting cancel will invalidate the One-Time Code"),
        ).toBeInTheDocument();
    });

    it("generates the code only once", async () => {
        const { rerender } = renderDialog();
        await waitFor(() => expect(generateMock).toHaveBeenCalledTimes(1));

        rerender(
            <IdentityVerificationDialog
                elevation={elevation}
                opening={true}
                handleClosed={vi.fn()}
                handleOpened={vi.fn()}
            />,
        );

        await waitFor(() => expect(generateMock).toHaveBeenCalledTimes(1));
    });
});

describe("code generation failures", () => {
    it("notifies and closes when generation rejects", async () => {
        generateMock.mockRejectedValue(new Error("boom"));

        const { handleClosed } = renderDialog();

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith(
                "Failed to generate the One-Time Code. Please try again later.",
            ),
        );
        expect(handleClosed).toHaveBeenCalledWith(false);
    });

    it("notifies and closes when generation returns nothing", async () => {
        generateMock.mockResolvedValue(null as any);

        const { handleClosed } = renderDialog();

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith(
                "Failed to generate the One-Time Code. Please try again later.",
            ),
        );
        expect(handleClosed).toHaveBeenCalledWith(false);
    });
});

describe("code entry", () => {
    it("records the typed code", async () => {
        renderDialog();
        await screen.findByText("Verify");

        fireEvent.change(getCode(), { target: { value: "123456" } });

        await waitFor(() => expect(getCode()).toHaveValue("123456"));
    });

    it("strips whitespace from the code", async () => {
        renderDialog();
        await screen.findByText("Verify");

        fireEvent.change(getCode(), { target: { value: "123 456" } });

        await waitFor(() => expect(getCode()).toHaveValue("123456"));
    });

    it("does nothing when verifying an empty code", async () => {
        renderDialog();
        await screen.findByText("Verify");

        clickVerify();

        await waitFor(() => expect(verifyMock).not.toHaveBeenCalled());
    });
});

describe("verification", () => {
    it("verifies the code and closes after the success delay", async () => {
        vi.useFakeTimers();

        const handleClosed = vi.fn();
        renderDialog({ handleClosed });

        await vi.waitFor(() => expect(screen.getByText("Verify")).toBeInTheDocument());
        fireEvent.change(getCode(), { target: { value: "123456" } });
        clickVerify();

        await vi.waitFor(() => expect(verifyMock).toHaveBeenCalledWith("123456"));
        await vi.waitFor(() => expect(screen.getByTestId("success-icon")).toBeInTheDocument());
        expect(handleClosed).not.toHaveBeenCalled();

        await act(async () => {
            await vi.advanceTimersByTimeAsync(750);
        });

        expect(handleClosed).toHaveBeenCalledWith(true);
    });

    it("hides the footer once verification succeeded", async () => {
        renderDialog();
        await screen.findByText("Verify");

        fireEvent.change(getCode(), { target: { value: "123456" } });
        clickVerify();

        await waitFor(() => expect(screen.getByTestId("success-icon")).toBeInTheDocument());
        expect(screen.queryByText("Verify")).not.toBeInTheDocument();
    });

    it("notifies, clears and refocuses the field on a wrong code", async () => {
        verifyMock.mockResolvedValue(false as any);

        renderDialog();
        await screen.findByText("Verify");

        fireEvent.change(getCode(), { target: { value: "123456" } });
        clickVerify();

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith(
                "The One-Time Code either doesn't match the one generated or an unknown error occurred",
            ),
        );
        await waitFor(() => expect(getCode()).toHaveValue(""));
        expect(getCode()).toHaveAttribute("aria-invalid", "true");
        expect(getCode()).toHaveFocus();
    });

    it("clears the error once the user retypes", async () => {
        verifyMock.mockResolvedValue(false as any);

        renderDialog();
        await screen.findByText("Verify");

        fireEvent.change(getCode(), { target: { value: "123456" } });
        clickVerify();
        await waitFor(() => expect(getCode()).toHaveAttribute("aria-invalid", "true"));

        fireEvent.change(getCode(), { target: { value: "654321" } });

        await waitFor(() => expect(getCode()).toHaveAttribute("aria-invalid", "false"));
    });

    it("disables the controls while verifying", async () => {
        let resolve: (value: unknown) => void = () => {};
        verifyMock.mockReturnValue(new Promise((r) => (resolve = r)) as any);

        renderDialog();
        await screen.findByText("Verify");

        fireEvent.change(getCode(), { target: { value: "123456" } });
        clickVerify();

        await waitFor(() => expect(document.getElementById("dialog-verify")).toBeDisabled());
        expect(document.getElementById("dialog-cancel")).toBeDisabled();
        expect(getCode()).toBeDisabled();

        resolve(true);
        await waitFor(() => expect(screen.getByTestId("success-icon")).toBeInTheDocument());
    });
});

describe("keyboard handling", () => {
    it("verifies on Enter", async () => {
        renderDialog();
        await screen.findByText("Verify");

        fireEvent.change(getCode(), { target: { value: "123456" } });
        fireEvent.keyDown(getCode(), { key: "Enter" });

        await waitFor(() => expect(verifyMock).toHaveBeenCalledWith("123456"));
    });

    it("flags an empty code on Enter", async () => {
        renderDialog();
        await screen.findByText("Verify");

        fireEvent.keyDown(getCode(), { key: "Enter" });

        await waitFor(() => expect(getCode()).toHaveAttribute("aria-invalid", "true"));
        expect(verifyMock).not.toHaveBeenCalled();
    });

    it("ignores keys other than Enter", async () => {
        renderDialog();
        await screen.findByText("Verify");

        fireEvent.change(getCode(), { target: { value: "123456" } });
        fireEvent.keyDown(getCode(), { key: "a" });

        expect(verifyMock).not.toHaveBeenCalled();
    });
});

describe("cancelling", () => {
    it("invalidates the code and closes", async () => {
        const { handleClosed } = renderDialog();
        await screen.findByText("Cancel");

        clickCancel();

        await waitFor(() => expect(deleteMock).toHaveBeenCalledWith("delete-id"));
        expect(handleClosed).toHaveBeenCalledWith(false);
    });

    it("logs failures invalidating the code", async () => {
        deleteMock.mockRejectedValue(new Error("boom"));

        const { handleClosed } = renderDialog();
        await screen.findByText("Cancel");

        clickCancel();

        await waitFor(() => expect(console.error).toHaveBeenCalledWith(expect.objectContaining({ message: "boom" })));
        expect(handleClosed).toHaveBeenCalledWith(false);
    });
});

describe("dismissal", () => {
    it("cancels when Escape is pressed", async () => {
        const { handleClosed } = renderDialog();
        await screen.findByText("Cancel");

        fireEvent.keyDown(document.activeElement ?? document.body, { key: "Escape" });

        await waitFor(() => expect(handleClosed).toHaveBeenCalledWith(false));
    });

    it("logs when there is no delete code to invalidate", async () => {
        generateMock.mockResolvedValue({} as any);

        const { handleClosed } = renderDialog();
        await screen.findByText("Cancel");

        clickCancel();

        await waitFor(() =>
            expect(console.error).toHaveBeenCalledWith(
                expect.objectContaining({ message: "The delete code was empty." }),
            ),
        );
        expect(deleteMock).not.toHaveBeenCalled();
        expect(handleClosed).toHaveBeenCalledWith(false);
    });
});

describe("cleanup", () => {
    it("clears the success timer on unmount", async () => {
        vi.useFakeTimers();

        const setTimeoutSpy = vi.spyOn(globalThis, "setTimeout");

        const { unmount } = renderDialog();

        await vi.waitFor(() => expect(screen.getByText("Verify")).toBeInTheDocument());
        fireEvent.change(getCode(), { target: { value: "123456" } });
        clickVerify();

        await vi.waitFor(() => expect(screen.getByTestId("success-icon")).toBeInTheDocument());

        const index = setTimeoutSpy.mock.calls.findIndex(([, timeout]) => timeout === 750);

        expect(index).toBeGreaterThanOrEqual(0);

        const timerID = setTimeoutSpy.mock.results[index].value;

        const clearTimeoutSpy = vi.spyOn(globalThis, "clearTimeout");

        unmount();

        expect(clearTimeoutSpy).toHaveBeenCalledWith(timerID);
    });
});
