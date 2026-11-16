// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { fireEvent, render, screen } from "@testing-library/react";

import WebAuthnTryIcon from "@components/WebAuthnTryIcon";
import { WebAuthnTouchState } from "@models/WebAuthn";

const mockTrigger = vi.fn();
const mockClear = vi.fn();
const mockOnRetryClick = vi.fn();

vi.mock("@hooks/Timer", () => ({
    useTimer: vi.fn(() => [50, mockTrigger, mockClear]),
}));

beforeEach(() => {
    mockTrigger.mockClear();
    mockClear.mockClear();
    mockOnRetryClick.mockClear();
});

vi.mock("@components/FailureIcon", () => ({
    default: () => <div data-testid="failure-icon" />,
}));

vi.mock("@components/FingerTouchIcon", () => ({
    default: () => <div data-testid="finger-touch" />,
}));

vi.mock("@components/UI/Spinner", () => ({
    Spinner: () => <div data-testid="spinner" />,
}));

vi.mock("@views/LoginPortal/SecondFactor/IconWithContext", () => ({
    default: ({
        children,
        className,
        icon,
    }: {
        icon: React.ReactNode;
        children: React.ReactNode;
        className?: string;
    }) => (
        <div className={className}>
            {icon}
            {children}
        </div>
    ),
}));

it("renders without crashing", () => {
    render(<WebAuthnTryIcon onRetryClick={mockOnRetryClick} webauthnTouchState={WebAuthnTouchState.WaitTouch} />);
});

it("displays touch icon when waiting", () => {
    render(<WebAuthnTryIcon onRetryClick={mockOnRetryClick} webauthnTouchState={WebAuthnTouchState.WaitTouch} />);
    const touchDiv = screen.getByTestId("finger-touch");
    expect(touchDiv.parentElement).not.toHaveClass("hidden");
    const failureDiv = screen.getByTestId("failure-icon");
    expect(failureDiv.parentElement).toHaveClass("hidden");
});

it("displays failure icon when failed", () => {
    render(<WebAuthnTryIcon onRetryClick={mockOnRetryClick} webauthnTouchState={WebAuthnTouchState.Failure} />);
    const touchDiv = screen.getByTestId("finger-touch");
    expect(touchDiv.parentElement).toHaveClass("hidden");
    const failureDiv = screen.getByTestId("failure-icon");
    expect(failureDiv.parentElement).not.toHaveClass("hidden");
});

it("retries on retry button click", () => {
    render(<WebAuthnTryIcon onRetryClick={mockOnRetryClick} webauthnTouchState={WebAuthnTouchState.Failure} />);
    const button = screen.getByRole("button");
    fireEvent.click(button);
    expect(mockOnRetryClick).toHaveBeenCalledTimes(1);
    expect(mockClear).toHaveBeenCalledTimes(1);
    expect(mockTrigger).toHaveBeenCalledTimes(2);
});

it("displays progress when in progress", () => {
    render(<WebAuthnTryIcon onRetryClick={mockOnRetryClick} webauthnTouchState={WebAuthnTouchState.InProgress} />);
    expect(screen.getByTestId("spinner").parentElement).not.toHaveClass("hidden");
    expect(screen.getByTestId("finger-touch").parentElement).toHaveClass("hidden");
    expect(screen.getByTestId("failure-icon").parentElement).toHaveClass("hidden");
});

it("displays failure without retry when unsupported", () => {
    render(<WebAuthnTryIcon onRetryClick={mockOnRetryClick} webauthnTouchState={WebAuthnTouchState.Unsupported} />);
    expect(screen.getByTestId("failure-icon").parentElement).not.toHaveClass("hidden");
    expect(screen.getByTestId("spinner").parentElement).toHaveClass("hidden");
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
});

it("offers another sign in method when unsupported", () => {
    const onMethodSelectionClick = vi.fn();

    render(
        <WebAuthnTryIcon
            onMethodSelectionClick={onMethodSelectionClick}
            onRetryClick={mockOnRetryClick}
            webauthnTouchState={WebAuthnTouchState.Unsupported}
        />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Use another sign in method" }));

    expect(onMethodSelectionClick).toHaveBeenCalledTimes(1);
    expect(mockOnRetryClick).not.toHaveBeenCalled();
});
