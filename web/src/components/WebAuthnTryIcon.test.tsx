import React from "react";

import { fireEvent, render, screen } from "@testing-library/react";
import { vi } from "vitest";

import WebAuthnTryIcon from "@components/WebAuthnTryIcon";
import { WebAuthnTouchState } from "@models/WebAuthn";

const mockTrigger = vi.fn();
const mockClear = vi.fn();
const mockOnRetryClick = vi.fn();

vi.mock("@hooks/Timer", () => ({
    useTimer: vi.fn(() => [50, mockTrigger, mockClear]),
}));

vi.mock("@components/FailureIcon", () => ({
    default: () => <div data-testid="failure-icon" />,
}));

vi.mock("@components/FingerTouchIcon", () => ({
    default: () => <div data-testid="finger-touch" />,
}));

vi.mock("@views/LoginPortal/SecondFactor/IconWithContext", () => ({
    default: ({
        icon,
        children,
        className,
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

it("shows touch icon when waiting", () => {
    render(<WebAuthnTryIcon onRetryClick={mockOnRetryClick} webauthnTouchState={WebAuthnTouchState.WaitTouch} />);
    const touchDiv = screen.getByTestId("finger-touch");
    expect(touchDiv.parentElement).not.toHaveClass("hidden");
    const failureDiv = screen.getByTestId("failure-icon");
    expect(failureDiv.parentElement).toHaveClass("hidden");
});

it("shows failure icon when failed", () => {
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
    expect(mockOnRetryClick).toHaveBeenCalled();
    expect(mockClear).toHaveBeenCalled();
    expect(mockTrigger).toHaveBeenCalled();
});
