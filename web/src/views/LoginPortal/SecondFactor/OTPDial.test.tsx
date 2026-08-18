import { render, screen } from "@testing-library/react";

import OTPDial, { State } from "@views/LoginPortal/SecondFactor/OTPDial";

vi.mock("@components/SuccessIcon", () => ({
    default: () => <div data-testid="success-icon" />,
}));

vi.mock("@components/TimerIcon", () => ({
    default: () => <div data-testid="timer-icon" />,
}));

vi.mock("@views/LoginPortal/SecondFactor/IconWithContext", () => ({
    default: (props: any) => (
        <div data-testid="icon-context">
            {props.icon}
            {props.children}
        </div>
    ),
}));

const inputs = () =>
    Array.from(document.querySelectorAll("#otp-input input")).filter(
        (el) => el.getAttribute("aria-hidden") !== "true",
    ) as HTMLInputElement[];

it("renders OTP input with correct digit count", () => {
    render(<OTPDial passcode="" state={State.Idle} digits={6} period={30} onChange={vi.fn()} />);
    expect(inputs()).toHaveLength(6);
});

it("renders timer icon in idle state", () => {
    render(<OTPDial passcode="" state={State.Idle} digits={6} period={30} onChange={vi.fn()} />);
    expect(screen.getByTestId("timer-icon")).toBeInTheDocument();
});

it("renders success icon in success state", () => {
    render(<OTPDial passcode="123456" state={State.Success} digits={6} period={30} onChange={vi.fn()} />);
    expect(screen.getByTestId("success-icon")).toBeInTheDocument();
});

it("disables input during in-progress state", () => {
    render(<OTPDial passcode="" state={State.InProgress} digits={6} period={30} onChange={vi.fn()} />);
    expect(inputs().every((el) => el.disabled)).toBe(true);
});

it("disables input during rate-limited state", () => {
    render(<OTPDial passcode="" state={State.RateLimited} digits={6} period={30} onChange={vi.fn()} />);
    expect(inputs().every((el) => el.disabled)).toBe(true);
});
