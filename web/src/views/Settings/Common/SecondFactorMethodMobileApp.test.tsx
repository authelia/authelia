// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { fireEvent, render, screen } from "@testing-library/react";

import SecondFactorMethodMobileApp from "@views/Settings/Common/SecondFactorMethodMobileApp";

const mocks = vi.hoisted(() => ({
    createErrorNotification: vi.fn(),
}));

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: mocks.createErrorNotification,
    }),
}));

vi.mock("@views/LoginPortal/SecondFactor/AppAuthenticatorChallenge", () => ({
    default: (props: any) => (
        <div data-testid="challenge" data-body={JSON.stringify(props.body)}>
            <button data-testid="challenge-success" onClick={() => props.onSuccess("https://example.com")} />
            <button data-testid="challenge-error" onClick={() => props.onError(new Error("app failed"))} />
        </div>
    ),
}));

beforeEach(() => {
    vi.clearAllMocks();
});

it("creates a challenge without a sign in target", () => {
    render(<SecondFactorMethodMobileApp onSecondFactorSuccess={vi.fn()} />);

    expect(screen.getByTestId("challenge")).toHaveAttribute("data-body", "{}");
});

it("reports success", () => {
    const onSecondFactorSuccess = vi.fn();

    render(<SecondFactorMethodMobileApp onSecondFactorSuccess={onSecondFactorSuccess} />);

    fireEvent.click(screen.getByTestId("challenge-success"));

    expect(onSecondFactorSuccess).toHaveBeenCalledOnce();
});

it("surfaces errors as notifications", () => {
    render(<SecondFactorMethodMobileApp onSecondFactorSuccess={vi.fn()} />);

    fireEvent.click(screen.getByTestId("challenge-error"));

    expect(mocks.createErrorNotification).toHaveBeenCalledWith("app failed");
});
