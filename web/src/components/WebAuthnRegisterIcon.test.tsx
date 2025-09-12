import React from "react";

import { render, screen } from "@testing-library/react";
import { vi } from "vitest";

import WebAuthnRegisterIcon from "@components/WebAuthnRegisterIcon";
import { useTimer } from "@hooks/Timer";

vi.mock("@hooks/Timer", () => ({
    useTimer: vi.fn(() => [50, vi.fn()]),
}));

vi.mock("@components/FingerTouchIcon", () => ({
    default: () => <div data-testid="finger-touch" />,
}));

it("renders without crashing", () => {
    render(<WebAuthnRegisterIcon timeout={30000} />);
});

it("renders finger touch icon", () => {
    render(<WebAuthnRegisterIcon timeout={30000} />);
    expect(screen.getByTestId("finger-touch")).toBeInTheDocument();
});

it("calls timer with timeout", () => {
    render(<WebAuthnRegisterIcon timeout={30000} />);
    expect(useTimer).toHaveBeenCalledWith(30000);
});
