import { fireEvent, render, screen } from "@testing-library/react";

import { SecondFactorMethod } from "@models/Methods";
import MethodSelectionDialog from "@views/LoginPortal/SecondFactor/MethodSelectionDialog";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@components/FingerTouchIcon", () => ({
    default: () => <div data-testid="finger-touch-icon" />,
}));

vi.mock("@components/PushNotificationIcon", () => ({
    default: () => <div data-testid="push-icon" />,
}));

vi.mock("@components/TimerIcon", () => ({
    default: () => <div data-testid="timer-icon" />,
}));

it("renders TOTP option when available", () => {
    render(
        <MethodSelectionDialog
            open={true}
            methods={new Set([SecondFactorMethod.TOTP])}
            webauthn={false}
            onClose={vi.fn()}
            onClick={vi.fn()}
        />,
    );
    expect(screen.getByText("Time-based One-Time Password")).toBeInTheDocument();
});

it("renders WebAuthn option when available and supported", () => {
    render(
        <MethodSelectionDialog
            open={true}
            methods={new Set([SecondFactorMethod.WebAuthn])}
            webauthn={true}
            onClose={vi.fn()}
            onClick={vi.fn()}
        />,
    );
    expect(screen.getByText("Security Key - WebAuthn")).toBeInTheDocument();
});

it("does not render WebAuthn option when not supported", () => {
    render(
        <MethodSelectionDialog
            open={true}
            methods={new Set([SecondFactorMethod.WebAuthn])}
            webauthn={false}
            onClose={vi.fn()}
            onClick={vi.fn()}
        />,
    );
    expect(screen.queryByText("Security Key - WebAuthn")).not.toBeInTheDocument();
});

it("renders Push Notification option when available", () => {
    render(
        <MethodSelectionDialog
            open={true}
            methods={new Set([SecondFactorMethod.MobilePush])}
            webauthn={false}
            onClose={vi.fn()}
            onClick={vi.fn()}
        />,
    );
    expect(screen.getByText("Push Notification")).toBeInTheDocument();
});

it("calls onClick with the correct method", () => {
    const onClick = vi.fn();
    render(
        <MethodSelectionDialog
            open={true}
            methods={new Set([SecondFactorMethod.TOTP])}
            webauthn={false}
            onClose={vi.fn()}
            onClick={onClick}
        />,
    );
    fireEvent.click(screen.getByText("Time-based One-Time Password"));
    expect(onClick).toHaveBeenCalledWith(SecondFactorMethod.TOTP);
});

it("calls onClose when Close button is clicked", () => {
    const onClose = vi.fn();
    render(
        <MethodSelectionDialog
            open={true}
            methods={new Set([SecondFactorMethod.TOTP])}
            webauthn={false}
            onClose={onClose}
            onClick={vi.fn()}
        />,
    );
    fireEvent.click(screen.getByText("Close"));
    expect(onClose).toHaveBeenCalledOnce();
});
