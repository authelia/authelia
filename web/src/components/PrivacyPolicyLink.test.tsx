import { render } from "@testing-library/react";

import PrivacyPolicyLink from "@components/PrivacyPolicyLink";

it("renders a link to the privacy policy with the correct text", () => {
    document.body.setAttribute("data-privacypolicyurl", "http://example.com/privacy-policy");

    const { getByRole } = render(<PrivacyPolicyLink />);
    const link = getByRole("link");
    expect(link).toHaveAttribute("href", "http://example.com/privacy-policy");
    expect(link).toHaveTextContent("Privacy Policy");
});
