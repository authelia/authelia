// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { fireEvent, render, screen } from "@testing-library/react";
import { expect, it, vi } from "vitest";

import ExternalIdentityLinkItem from "@views/Settings/ExternalIdentity/ExternalIdentityLinkItem";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

const problem =
    "This link could not be verified and can't be used to sign in, remove it to link your {{name}} account again";

const link = {
    created_at: "2026-09-01T00:00:00Z",
    id: 7,
    issuer: "https://accounts.google.com",
    provider: "google",
    provider_name: "Google",
    remote_username: "john",
    subject: "abc123",
    type: "openid_connect",
};

it("shows a verified link without a problem", () => {
    render(<ExternalIdentityLinkItem link={link} onDelete={vi.fn()} />);

    expect(screen.getByText("Google")).toBeInTheDocument();
    expect(screen.getByText("(john)")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: problem })).not.toBeInTheDocument();
});

it("marks a link which could not be verified and hides its remote username", () => {
    render(<ExternalIdentityLinkItem link={{ ...link, invalid: true }} onDelete={vi.fn()} />);

    expect(screen.getByText("Google")).toBeInTheDocument();
    expect(screen.queryByText("(john)")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: problem })).toBeInTheDocument();
});

it("allows a link which could not be verified to be deleted", () => {
    const onDelete = vi.fn();

    render(<ExternalIdentityLinkItem link={{ ...link, invalid: true }} onDelete={onDelete} />);

    fireEvent.click(screen.getByRole("button", { name: "Remove the link to your {{name}} account" }));

    expect(onDelete).toHaveBeenCalledOnce();
});
