// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { fireEvent, render, screen } from "@testing-library/react";
import { expect, it, vi } from "vitest";

import ExternalIdentityLinkPendingPanel from "@views/Settings/ExternalIdentity/ExternalIdentityLinkPendingPanel";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

const pending = {
    email: "john@example.com",
    issuer: "https://accounts.google.com",
    provider: "google",
    provider_name: "Google",
    remote_username: "john",
    subject: "abc123",
};

it("summarizes the proposed link", () => {
    render(<ExternalIdentityLinkPendingPanel pending={pending} onAccept={vi.fn()} onDecline={vi.fn()} />);

    expect(screen.getByText("Google")).toBeInTheDocument();
    expect(screen.getByText("john")).toBeInTheDocument();
    expect(screen.getByText("john@example.com")).toBeInTheDocument();
    expect(screen.getByText("abc123")).toBeInTheDocument();
});

it("prefers the display name over the remote username when both are present", () => {
    const pendingWithDisplayName = { ...pending, display_name: "John Doe" };

    render(
        <ExternalIdentityLinkPendingPanel pending={pendingWithDisplayName} onAccept={vi.fn()} onDecline={vi.fn()} />,
    );

    expect(screen.getByText("John Doe")).toBeInTheDocument();
    expect(screen.queryByText("john")).not.toBeInTheDocument();
});

it("calls onAccept when accepted", () => {
    const onAccept = vi.fn();

    render(<ExternalIdentityLinkPendingPanel pending={pending} onAccept={onAccept} onDecline={vi.fn()} />);

    fireEvent.click(screen.getByRole("button", { name: "Accept" }));

    expect(onAccept).toHaveBeenCalledOnce();
});

it("calls onDecline when declined", () => {
    const onDecline = vi.fn();

    render(<ExternalIdentityLinkPendingPanel pending={pending} onAccept={vi.fn()} onDecline={onDecline} />);

    fireEvent.click(screen.getByRole("button", { name: "Decline" }));

    expect(onDecline).toHaveBeenCalledOnce();
});
