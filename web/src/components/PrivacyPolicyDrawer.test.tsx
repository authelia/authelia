import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach } from "vitest";

import PrivacyPolicyDrawer from "@components/PrivacyPolicyDrawer";

beforeEach(() => {
    document.body.setAttribute("data-privacypolicyurl", "");
    document.body.setAttribute("data-privacypolicyaccept", "false");

    global.localStorage.clear();
});

it("renders privacy policy and accepts when Accept button is clicked", () => {
    document.body.setAttribute("data-privacypolicyurl", "http://example.com/privacy-policy");
    document.body.setAttribute("data-privacypolicyaccept", "true");

    const { container } = render(<PrivacyPolicyDrawer />);
    fireEvent.click(screen.getByText("Accept"));
    expect(container).toBeEmptyDOMElement();
});

it("does not render when privacy policy is disabled", () => {
    render(<PrivacyPolicyDrawer />);
    expect(screen.queryByText("Privacy Policy")).toBeNull();
    expect(screen.queryByText("You must view and accept the Privacy Policy before using")).toBeNull();
    expect(screen.queryByText("Accept")).toBeNull();
});

it("does not render when acceptance is not required", () => {
    document.body.setAttribute("data-privacypolicyurl", "http://example.com/privacy-policy");

    render(<PrivacyPolicyDrawer />);
    expect(screen.queryByText("Privacy Policy")).toBeNull();
    expect(screen.queryByText("You must view and accept the Privacy Policy before using")).toBeNull();
    expect(screen.queryByText("Accept")).toBeNull();
});

it("does not render when already accepted", () => {
    global.localStorage.setItem("privacy-policy-accepted", "true");

    const { container } = render(<PrivacyPolicyDrawer />);
    expect(container).toBeEmptyDOMElement();
});
