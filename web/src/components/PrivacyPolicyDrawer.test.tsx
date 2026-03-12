import { fireEvent, render, screen } from "@testing-library/react";

import PrivacyPolicyDrawer from "@components/PrivacyPolicyDrawer";

beforeEach(() => {
    document.body.dataset.privacypolicyurl = "";
    document.body.dataset.privacypolicyaccept = "false";

    globalThis.localStorage.clear();
});

it("renders privacy policy and accepts when Accept button is clicked", () => {
    document.body.dataset.privacypolicyurl = "http://example.com/privacy-policy";
    document.body.dataset.privacypolicyaccept = "true";

    const { container } = render(<PrivacyPolicyDrawer />);
    fireEvent.click(screen.getByText("Accept"));
    expect(container).toBeEmptyDOMElement();
});

it("does not render when privacy policy is disabled", () => {
    render(<PrivacyPolicyDrawer />);
    expect(screen.queryByText("Privacy Policy")).not.toBeInTheDocument();
    expect(screen.queryByText("You must view and accept the Privacy Policy before using")).not.toBeInTheDocument();
    expect(screen.queryByText("Accept")).not.toBeInTheDocument();
});

it("does not render when acceptance is not required", () => {
    document.body.dataset.privacypolicyurl = "http://example.com/privacy-policy";

    render(<PrivacyPolicyDrawer />);
    expect(screen.queryByText("Privacy Policy")).not.toBeInTheDocument();
    expect(screen.queryByText("You must view and accept the Privacy Policy before using")).not.toBeInTheDocument();
    expect(screen.queryByText("Accept")).not.toBeInTheDocument();
});

it("does not render when already accepted", () => {
    globalThis.localStorage.setItem("privacy-policy-accepted", "true");

    const { container } = render(<PrivacyPolicyDrawer />);
    expect(container).toBeEmptyDOMElement();
});
