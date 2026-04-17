import { act, render, screen } from "@testing-library/react";

import SecondFactorMethodWebAuthn from "@views/Settings/Common/SecondFactorMethodWebAuthn";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@models/WebAuthn", () => ({
    AssertionResult: { Success: 0 },
    AssertionResultFailureString: () => "failure",
    WebAuthnTouchState: { Failure: 2, InProgress: 1, WaitTouch: 0 },
}));

vi.mock("@services/WebAuthn", () => ({
    getWebAuthnOptions: vi.fn().mockResolvedValue({ options: null, status: 400 }),
    getWebAuthnResult: vi.fn(),
    postWebAuthnResponse: vi.fn(),
}));

vi.mock("@components/WebAuthnTryIcon", () => ({
    default: (props: any) => <div data-testid="webauthn-try-icon" data-state={props.webauthnTouchState} />,
}));

it("renders WebAuthnTryIcon", async () => {
    vi.spyOn(console, "error").mockImplementation(() => {});
    await act(async () => {
        render(<SecondFactorMethodWebAuthn onSecondFactorSuccess={vi.fn()} />);
    });
    expect(screen.getByTestId("webauthn-try-icon")).toBeInTheDocument();
});
