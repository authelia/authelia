import React from "react";

import { render, screen } from "@testing-library/react";

import { ScopeAvatar, ScopeDescription } from "@components/OpenIDConnect";

it("returns correct avatar for openid", () => {
    render(ScopeAvatar("openid"));
    expect(screen.getByTestId("AccountBoxIcon")).toBeInTheDocument();
});

it("returns correct avatar for offline_access", () => {
    render(ScopeAvatar("offline_access"));
    expect(screen.getByTestId("AutorenewIcon")).toBeInTheDocument();
});

it("returns correct avatar for profile", () => {
    render(ScopeAvatar("profile"));
    expect(screen.getByTestId("ContactsIcon")).toBeInTheDocument();
});

it("returns correct avatar for groups", () => {
    render(ScopeAvatar("groups"));
    expect(screen.getByTestId("GroupIcon")).toBeInTheDocument();
});

it("returns correct avatar for email", () => {
    render(ScopeAvatar("email"));
    expect(screen.getByTestId("DraftsIcon")).toBeInTheDocument();
});

it("returns correct avatar for phone", () => {
    render(ScopeAvatar("phone"));
    expect(screen.getByTestId("PhoneAndroidIcon")).toBeInTheDocument();
});

it("returns correct avatar for address", () => {
    render(ScopeAvatar("address"));
    expect(screen.getByTestId("HomeIcon")).toBeInTheDocument();
});

it("returns correct avatar for authelia.bearer.authz", () => {
    render(ScopeAvatar("authelia.bearer.authz"));
    expect(screen.getByTestId("LockOpenIcon")).toBeInTheDocument();
});

it("returns policy avatar for unknown scope", () => {
    render(ScopeAvatar("unknown"));
    expect(screen.getByTestId("PolicyIcon")).toBeInTheDocument();
});

it("returns correct description for openid", () => {
    expect(ScopeDescription("openid")).toBe("Use OpenID to verify your identity");
});

it("returns correct description for offline_access", () => {
    expect(ScopeDescription("offline_access")).toBe("Automatically refresh these permissions without user interaction");
});

it("returns correct description for profile", () => {
    expect(ScopeDescription("profile")).toBe("Access your profile information");
});

it("returns correct description for groups", () => {
    expect(ScopeDescription("groups")).toBe("Access your group membership");
});

it("returns correct description for email", () => {
    expect(ScopeDescription("email")).toBe("Access your email addresses");
});

it("returns correct description for phone", () => {
    expect(ScopeDescription("phone")).toBe("Access your phone number");
});

it("returns correct description for address", () => {
    expect(ScopeDescription("address")).toBe("Access your address");
});

it("returns correct description for authelia.bearer.authz", () => {
    expect(ScopeDescription("authelia.bearer.authz")).toBe("Access protected resources logged in as you");
});

it("returns scope for unknown description", () => {
    expect(ScopeDescription("unknown")).toBe("unknown");
});
