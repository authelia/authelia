import { render, screen } from "@testing-library/react";

import SecondFactorMethodMobilePush from "@views/Settings/Common/SecondFactorMethodMobilePush";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("tss-react/mui", () => ({
    makeStyles: () => () => () => ({
        classes: { container: "", icon: "" },
        cx: (...args: any[]) => args.filter(Boolean).join(" "),
    }),
}));

vi.mock("@hooks/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: vi.fn(),
    }),
}));

vi.mock("@services/PushNotification", () => ({
    completeDuoDeviceSelectionProcess: vi.fn(),
    completePushNotificationSignIn: vi.fn().mockResolvedValue(null),
    initiateDuoDeviceSelectionProcess: vi.fn(),
}));

vi.mock("@components/FailureIcon", () => ({
    default: () => <div data-testid="failure-icon" />,
}));

vi.mock("@components/PushNotificationIcon", () => ({
    default: () => <div data-testid="push-notification-icon" />,
}));

vi.mock("@views/LoginPortal/SecondFactor/DeviceSelectionContainer", () => ({
    default: () => <div data-testid="device-selection" />,
}));

it("renders push notification icon", () => {
    vi.spyOn(console, "error").mockImplementation(() => {});
    render(<SecondFactorMethodMobilePush onSecondFactorSuccess={vi.fn()} />);
    expect(screen.getByTestId("push-notification-icon")).toBeInTheDocument();
});

it("renders select a device link", () => {
    vi.spyOn(console, "error").mockImplementation(() => {});
    render(<SecondFactorMethodMobilePush onSecondFactorSuccess={vi.fn()} />);
    expect(screen.getByText("Select a Device")).toBeInTheDocument();
});
