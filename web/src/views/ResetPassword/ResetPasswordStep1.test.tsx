import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";

import ResetPasswordStep1 from "@views/ResetPassword/ResetPasswordStep1";

const mockNavigate = vi.fn();

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("react-router", () => ({
    useNavigate: () => mockNavigate,
}));

vi.mock("@components/ComponentWithTooltip", () => ({
    default: (props: any) => <div>{props.children}</div>,
}));

vi.mock("@constants/Routes", () => ({
    IndexRoute: "/",
}));

const mockCreateError = vi.fn();
const mockCreateInfo = vi.fn();

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: mockCreateError,
        createInfoNotification: mockCreateInfo,
    }),
}));

vi.mock("@layouts/MinimalLayout", () => ({
    default: (props: any) => <div data-testid="layout">{props.children}</div>,
}));

vi.mock("@services/ResetPassword", () => ({
    initiateResetPasswordProcess: vi.fn(),
}));

beforeEach(() => {
    mockNavigate.mockReset();
    mockCreateError.mockReset();
    mockCreateInfo.mockReset();
});

afterEach(() => {
    vi.useRealTimers();
});

it("renders the reset password form", () => {
    render(<ResetPasswordStep1 />);
    expect(screen.getByLabelText("Username")).toBeInTheDocument();
    expect(screen.getByText("Reset")).toBeInTheDocument();
    expect(screen.getByText("Cancel")).toBeInTheDocument();
});

it("shows error when username is empty", async () => {
    render(<ResetPasswordStep1 />);

    await act(async () => {
        fireEvent.click(screen.getByText("Reset"));
    });

    expect(mockCreateError).toHaveBeenCalledWith("Username is required");
});

it("navigates to index on successful reset", async () => {
    const { initiateResetPasswordProcess } = await import("@services/ResetPassword");
    vi.mocked(initiateResetPasswordProcess).mockResolvedValue({ limited: false, retryAfter: 0 });

    render(<ResetPasswordStep1 />);

    await act(async () => {
        fireEvent.change(screen.getByLabelText("Username"), { target: { value: "testuser" } });
    });

    await act(async () => {
        fireEvent.click(screen.getByText("Reset"));
    });

    expect(mockCreateInfo).toHaveBeenCalled();
    expect(mockNavigate).toHaveBeenCalledWith("/");
});

it("shows rate limited error", async () => {
    const { initiateResetPasswordProcess } = await import("@services/ResetPassword");
    vi.mocked(initiateResetPasswordProcess).mockResolvedValue({ limited: true, retryAfter: 30 });

    render(<ResetPasswordStep1 />);

    await act(async () => {
        fireEvent.change(screen.getByLabelText("Username"), { target: { value: "testuser" } });
    });

    await act(async () => {
        fireEvent.click(screen.getByText("Reset"));
    });

    expect(mockCreateError).toHaveBeenCalledWith("You have made too many requests");
});

it("navigates to index when cancel is clicked", () => {
    render(<ResetPasswordStep1 />);
    fireEvent.click(screen.getByText("Cancel"));
    expect(mockNavigate).toHaveBeenCalledWith("/");
});

it("shows error when service throws", async () => {
    const { initiateResetPasswordProcess } = await import("@services/ResetPassword");
    vi.mocked(initiateResetPasswordProcess).mockRejectedValue(new Error("network error"));

    render(<ResetPasswordStep1 />);

    await act(async () => {
        fireEvent.change(screen.getByLabelText("Username"), { target: { value: "testuser" } });
    });

    await act(async () => {
        fireEvent.click(screen.getByText("Reset"));
    });

    expect(mockCreateError).toHaveBeenCalledWith("There was an issue initiating the password reset process");
});

describe("rate limiting", () => {
    it("re-enables the reset button after the retry window", async () => {
        vi.useFakeTimers();
        const { initiateResetPasswordProcess } = await import("@services/ResetPassword");
        vi.mocked(initiateResetPasswordProcess).mockResolvedValue({ limited: true, retryAfter: 30 });

        render(<ResetPasswordStep1 />);

        fireEvent.change(screen.getByLabelText("Username"), { target: { value: "john" } });

        await act(async () => {
            fireEvent.click(screen.getByText("Reset"));
        });

        expect(mockCreateError).toHaveBeenCalledWith("You have made too many requests");
        expect(document.getElementById("reset-button")).toBeDisabled();

        await act(async () => {
            await vi.advanceTimersByTimeAsync(30000);
        });

        expect(document.getElementById("reset-button")).not.toBeDisabled();
    });

    it("replaces an in flight rate limit timer", async () => {
        vi.useFakeTimers();
        const { initiateResetPasswordProcess } = await import("@services/ResetPassword");
        vi.mocked(initiateResetPasswordProcess).mockResolvedValue({ limited: true, retryAfter: 30 });

        render(<ResetPasswordStep1 />);

        fireEvent.change(screen.getByLabelText("Username"), { target: { value: "john" } });

        await act(async () => {
            fireEvent.keyDown(screen.getByLabelText("Username"), { key: "Enter" });
        });

        expect(initiateResetPasswordProcess).toHaveBeenCalledTimes(1);
        expect(document.getElementById("reset-button")).toBeDisabled();

        // Resubmit while the first retry window is still open, so its timer is replaced rather than expired.
        await act(async () => {
            await vi.advanceTimersByTimeAsync(20000);
        });

        await act(async () => {
            fireEvent.keyDown(screen.getByLabelText("Username"), { key: "Enter" });
        });

        expect(initiateResetPasswordProcess).toHaveBeenCalledTimes(2);

        // Reaching the original deadline must not re-enable the button now that the timer has been replaced.
        await act(async () => {
            await vi.advanceTimersByTimeAsync(10000);
        });

        expect(document.getElementById("reset-button")).toBeDisabled();

        // Only the replacement window re-enables it.
        await act(async () => {
            await vi.advanceTimersByTimeAsync(20000);
        });

        expect(document.getElementById("reset-button")).not.toBeDisabled();
    });

    it("clears the rate limit timer on unmount", async () => {
        vi.useFakeTimers();
        const { initiateResetPasswordProcess } = await import("@services/ResetPassword");
        vi.mocked(initiateResetPasswordProcess).mockResolvedValue({ limited: true, retryAfter: 30 });

        const { unmount } = render(<ResetPasswordStep1 />);

        fireEvent.change(screen.getByLabelText("Username"), { target: { value: "john" } });

        await act(async () => {
            fireEvent.click(screen.getByText("Reset"));
        });

        expect(document.getElementById("reset-button")).toBeDisabled();

        const clearTimeoutSpy = vi.spyOn(globalThis, "clearTimeout");

        unmount();

        expect(clearTimeoutSpy).toHaveBeenCalled();
    });
});

describe("keyboard handling", () => {
    it("submits on Enter", async () => {
        const { initiateResetPasswordProcess } = await import("@services/ResetPassword");
        vi.mocked(initiateResetPasswordProcess).mockResolvedValue({ limited: false, retryAfter: 0 });

        render(<ResetPasswordStep1 />);

        fireEvent.change(screen.getByLabelText("Username"), { target: { value: "john" } });
        fireEvent.keyDown(screen.getByLabelText("Username"), { key: "Enter" });

        await waitFor(() => expect(initiateResetPasswordProcess).toHaveBeenCalledWith("john"));
    });

    it("ignores keys other than Enter", async () => {
        const { initiateResetPasswordProcess } = await import("@services/ResetPassword");
        vi.mocked(initiateResetPasswordProcess).mockReset();

        render(<ResetPasswordStep1 />);

        fireEvent.change(screen.getByLabelText("Username"), { target: { value: "john" } });
        fireEvent.keyDown(screen.getByLabelText("Username"), { key: "a" });

        expect(initiateResetPasswordProcess).not.toHaveBeenCalled();
    });
});

describe("responses", () => {
    it("reports a missing response", async () => {
        const { initiateResetPasswordProcess } = await import("@services/ResetPassword");
        vi.mocked(initiateResetPasswordProcess).mockResolvedValue(undefined as any);

        render(<ResetPasswordStep1 />);

        fireEvent.change(screen.getByLabelText("Username"), { target: { value: "john" } });
        fireEvent.click(screen.getByText("Reset"));

        await waitFor(() =>
            expect(mockCreateError).toHaveBeenCalledWith("There was an issue initiating the password reset process"),
        );
    });

    it("disables the form while loading", async () => {
        const { initiateResetPasswordProcess } = await import("@services/ResetPassword");
        let resolve: (value: unknown) => void = () => {};
        vi.mocked(initiateResetPasswordProcess).mockReturnValue(new Promise((r) => (resolve = r)) as any);

        render(<ResetPasswordStep1 />);

        fireEvent.change(screen.getByLabelText("Username"), { target: { value: "john" } });
        fireEvent.click(screen.getByText("Reset"));

        await waitFor(() => expect(screen.getByLabelText("Username")).toBeDisabled());
        expect(document.getElementById("cancel-button")).toBeDisabled();

        resolve({ limited: false, retryAfter: 0 });
        await waitFor(() => expect(screen.getByLabelText("Username")).not.toBeDisabled());
    });
});
