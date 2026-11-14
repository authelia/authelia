// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { fireEvent, render, screen } from "@testing-library/react";

import AppAuthenticatorDeviceItem from "@views/Settings/TwoFactorAuthentication/AppAuthenticatorDeviceItem";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({
        t: (key: string, opts?: Record<string, string>) =>
            opts ? key.replace(/{{(\w+)}}/g, (_, name: string) => opts[name] ?? "") : key,
    }),
}));

vi.mock("@hooks/RelativeTimeString", () => ({
    useRelativeTime: (date?: Date) => (date ? `relative ${date.toISOString()}` : ""),
}));

const device = {
    attestation_format: "none",
    created_at: "2026-11-12T10:00:00Z",
    id: 7,
    model: "iPhone15,2",
    name: "Jane's iPhone",
    platform: "iOS",
};

function renderItem(overrides: Record<string, unknown> = {}) {
    const handlers = { handleDelete: vi.fn(), handleEdit: vi.fn() };

    render(<AppAuthenticatorDeviceItem index={3} device={{ ...device, ...overrides } as any} {...handlers} />);

    return handlers;
}

it("renders the name, model, platform and dates", () => {
    renderItem({ last_used_at: "2026-11-12T11:00:00Z" });

    expect(screen.getByText("Jane's iPhone")).toBeInTheDocument();
    expect(screen.getByText("iPhone15,2, iOS")).toBeInTheDocument();
    expect(screen.getByText("Added relative 2026-11-12T10:00:00.000Z")).toBeInTheDocument();
    expect(screen.getByText("Last Used relative 2026-11-12T11:00:00.000Z")).toBeInTheDocument();
});

it("renders never used without a last used time", () => {
    renderItem();

    expect(screen.getByText("Never used")).toBeInTheDocument();
});

it("renders not attested for format none", () => {
    renderItem();

    expect(screen.getByText("Not attested")).toBeInTheDocument();
});

it("renders the verified identity", () => {
    renderItem({ attestation_format: "apple-appattest", attestation_identity: "ABCDE12345.com.example.app" });

    expect(screen.getByText("Verified: ABCDE12345.com.example.app")).toBeInTheDocument();
});

it("renders the official app", () => {
    renderItem({ attestation_format: "android-key", attestation_identity: "com.authelia.mobile" });

    expect(screen.getByText("Verified Authelia app")).toBeInTheDocument();
});

it("passes its index to the edit and delete handlers", () => {
    const handlers = renderItem();

    fireEvent.click(document.getElementById("app-authenticator-device-3-edit")!);
    fireEvent.click(document.getElementById("app-authenticator-device-3-delete")!);

    expect(handlers.handleEdit).toHaveBeenCalledWith(3);
    expect(handlers.handleDelete).toHaveBeenCalledWith(3);
});
