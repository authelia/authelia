import { render, screen } from "@testing-library/react";

import { AuthenticationLevel } from "@services/State";
import PushNotificationMethod from "@views/LoginPortal/SecondFactor/PushNotificationMethod";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@hooks/QueryParam", () => ({
    useQueryParam: () => null,
}));

vi.mock("@hooks/Flow", () => ({
    useFlow: () => ({ flow: null, id: null, subflow: null }),
    useFlowPresent: () => false,
}));

vi.mock("@hooks/OpenIDConnect", () => ({
    useUserCode: () => null,
}));

vi.mock("@services/PushNotification", () => ({
    completeDuoDeviceSelectionProcess: vi.fn(),
    completePushNotificationSignIn: vi.fn().mockResolvedValue(null),
    getPreferredDuoDevice: vi.fn().mockResolvedValue({}),
    initiateDuoDeviceSelectionProcess: vi.fn(),
}));

vi.mock("@components/FailureIcon", () => ({
    default: () => <div data-testid="failure-icon" />,
}));

vi.mock("@components/PushNotificationIcon", () => ({
    default: () => <div data-testid="push-notification-icon" />,
}));

vi.mock("@components/SuccessIcon", () => ({
    default: () => <div data-testid="success-icon" />,
}));

vi.mock("@views/LoginPortal/SecondFactor/MethodContainer", () => ({
    default: (props: any) => (
        <div data-testid="method-container" data-state={props.state} data-title={props.title}>
            {props.children}
        </div>
    ),
    State: { ALREADY_AUTHENTICATED: "ALREADY_AUTHENTICATED", METHOD: "METHOD", NOT_REGISTERED: "NOT_REGISTERED" },
}));

vi.mock("@views/LoginPortal/SecondFactor/DeviceSelectionContainer", () => ({
    default: () => <div data-testid="device-selection" />,
}));

afterEach(() => {
    vi.restoreAllMocks();
});

it("renders method container with push notification title", () => {
    vi.spyOn(console, "error").mockImplementation(() => {});
    vi.spyOn(console, "debug").mockImplementation(() => {});
    render(
        <PushNotificationMethod
            id="push-method"
            authenticationLevel={AuthenticationLevel.OneFactor}
            duoSelfEnrollment={false}
            registered={true}
            onSignInError={vi.fn()}
            onSelectionClick={vi.fn()}
            onSignInSuccess={vi.fn()}
        />,
    );
    expect(screen.getByTestId("method-container")).toHaveAttribute("data-title", "Push Notification");
});

it("renders already authenticated state at two factor level", () => {
    vi.spyOn(console, "debug").mockImplementation(() => {});
    render(
        <PushNotificationMethod
            id="push-method"
            authenticationLevel={AuthenticationLevel.TwoFactor}
            duoSelfEnrollment={false}
            registered={true}
            onSignInError={vi.fn()}
            onSelectionClick={vi.fn()}
            onSignInSuccess={vi.fn()}
        />,
    );
    expect(screen.getByTestId("method-container")).toHaveAttribute("data-state", "ALREADY_AUTHENTICATED");
});
