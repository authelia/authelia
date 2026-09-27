// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { render, screen } from "@testing-library/react";

import WebAuthnCredentialsDisabledPanel from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialsDisabledPanel";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

it("renders the disabled message", () => {
    render(<WebAuthnCredentialsDisabledPanel />);
    expect(screen.getByText("WebAuthn Credentials")).toBeInTheDocument();
});
