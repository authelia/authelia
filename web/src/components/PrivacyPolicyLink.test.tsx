import { render, screen } from "@testing-library/react";

import PrivacyPolicyLink from "@components/PrivacyPolicyLink";

it("renders a link to the privacy policy with the correct text", () => {
    document.body.dataset.privacypolicyurl = "http://example.com/privacy-policy";

    render(<PrivacyPolicyLink />);
    const link = screen.getByRole("link");
    expect(link).toHaveAttribute("href", "http://example.com/privacy-policy");
    expect(link).toHaveTextContent("Privacy Policy");
});
