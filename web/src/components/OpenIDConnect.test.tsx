import { render } from "@testing-library/react";

import { ScopeAvatar, ScopeDescription } from "@components/OpenIDConnect";

function expectLucideIcon(container: HTMLElement, iconClass: string) {
    expect(container.querySelector(`svg.lucide-${iconClass}`)).toBeInTheDocument();
}

it("returns correct avatar for openid", () => {
    const { container } = render(ScopeAvatar("openid"));
    expectLucideIcon(container, "circle-user-round");
});

it("returns correct avatar for offline_access", () => {
    const { container } = render(ScopeAvatar("offline_access"));
    expectLucideIcon(container, "refresh-cw");
});

it("returns correct avatar for profile", () => {
    const { container } = render(ScopeAvatar("profile"));
    expectLucideIcon(container, "user-round");
});

it("returns correct avatar for groups", () => {
    const { container } = render(ScopeAvatar("groups"));
    expectLucideIcon(container, "users");
});

it("returns correct avatar for email", () => {
    const { container } = render(ScopeAvatar("email"));
    expectLucideIcon(container, "mail");
});

it("returns correct avatar for phone", () => {
    const { container } = render(ScopeAvatar("phone"));
    expectLucideIcon(container, "phone");
});

it("returns correct avatar for address", () => {
    const { container } = render(ScopeAvatar("address"));
    expectLucideIcon(container, "house");
});

it("returns correct avatar for authelia.bearer.authz", () => {
    const { container } = render(ScopeAvatar("authelia.bearer.authz"));
    expectLucideIcon(container, "lock");
});

it("returns policy avatar for unknown scope", () => {
    const { container } = render(ScopeAvatar("unknown"));
    expectLucideIcon(container, "shield");
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
