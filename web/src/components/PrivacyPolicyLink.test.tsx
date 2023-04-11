import React from "react";

import { render } from "@testing-library/react";

import PrivacyPolicyLink from "@components/PrivacyPolicyLink";

vi.mock("react-i18next", () => ({
    withTranslation: () => (Component: any) => {
        Component.defaultProps = { ...Component.defaultProps, t: (children: any) => children };
        return Component;
    },
    Trans: ({ children }: any) => children,
    useTranslation: () => {
        return {
            t: (str) => str,
            i18n: {
                changeLanguage: () => new Promise(() => {}),
            },
        };
    },
}));

it("renders a link to the privacy policy with the correct text", () => {
    document.body.setAttribute("data-privacypolicyurl", "http://example.com/privacy-policy");

    const { getByRole } = render(<PrivacyPolicyLink />);
    const link = getByRole("link");
    expect(link).toHaveAttribute("href", "http://example.com/privacy-policy");
    expect(link).toHaveTextContent("Privacy Policy");
});
