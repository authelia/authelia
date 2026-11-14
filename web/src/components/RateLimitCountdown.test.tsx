// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { render, screen } from "@testing-library/react";

import RateLimitCountdown from "@components/RateLimitCountdown";

const mocks = vi.hoisted(() => ({
    translate: vi.fn((key: string, options?: { count?: number }) => `${key}:${options?.count}`),
}));

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: mocks.translate }),
}));

it("renders only an empty status region when there are no seconds remaining", () => {
    render(<RateLimitCountdown secondsRemaining={0} />);

    expect(screen.getByRole("status")).toBeEmptyDOMElement();
    expect(document.getElementById("rate-limit-countdown")).not.toBeInTheDocument();
});

it("renders the seconds remaining", () => {
    render(<RateLimitCountdown secondsRemaining={42} />);

    expect(document.getElementById("rate-limit-countdown")).toHaveTextContent("Try again in {{count}} seconds:42");
    expect(mocks.translate).toHaveBeenCalledWith("Try again in {{count}} seconds", { count: 42 });
});

it("announces the seconds remaining once when the countdown starts", () => {
    const { rerender } = render(<RateLimitCountdown secondsRemaining={0} />);

    rerender(<RateLimitCountdown secondsRemaining={30} />);

    expect(screen.getByRole("status")).toHaveTextContent("Try again in {{count}} seconds:30");

    rerender(<RateLimitCountdown secondsRemaining={29} />);

    expect(screen.getByRole("status")).toHaveTextContent("Try again in {{count}} seconds:30");
    expect(document.getElementById("rate-limit-countdown")).toHaveTextContent("Try again in {{count}} seconds:29");
});

it("announces again when the countdown restarts", () => {
    const { rerender } = render(<RateLimitCountdown secondsRemaining={10} />);

    rerender(<RateLimitCountdown secondsRemaining={5} />);
    rerender(<RateLimitCountdown secondsRemaining={20} />);

    expect(screen.getByRole("status")).toHaveTextContent("Try again in {{count}} seconds:20");
});

it("clears the announcement when the countdown ends", () => {
    const { rerender } = render(<RateLimitCountdown secondsRemaining={1} />);

    rerender(<RateLimitCountdown secondsRemaining={0} />);

    expect(screen.getByRole("status")).toBeEmptyDOMElement();
    expect(document.getElementById("rate-limit-countdown")).not.toBeInTheDocument();
});
