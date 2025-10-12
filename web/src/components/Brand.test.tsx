import React from "react";

import { render, screen } from "@testing-library/react";
import { vi } from "vitest";

import Brand from "@components/Brand";

vi.mock("@constants/constants", () => ({
    EncodedName: Array.from("Authelia".split("").map((c) => c.charCodeAt(0))),
    EncodedURL: Array.from("https://www.authelia.com".split("").map((c) => c.charCodeAt(0))),
}));

vi.stubGlobal(
    "atob",
    vi.fn((str) => str),
);

it("renders without crashing", () => {
    render(<Brand />);
});

it("renders link with correct text", () => {
    render(<Brand />);
    const link = screen.getByRole("link");
    expect(link).toHaveAttribute("href", "https://www.authelia.com");
    expect(link).toHaveTextContent("Powered by Authelia");
});

it("renders privacy policy when enabled", () => {
    document.body.setAttribute("data-privacypolicyurl", "https://example.com");
    render(<Brand />);
    expect(screen.getByText("Privacy Policy")).toBeInTheDocument();
});

it("does not render privacy policy when disabled", () => {
    document.body.setAttribute("data-privacypolicyurl", "");
    render(<Brand />);
    expect(screen.queryByText("Privacy Policy")).toBeNull();
});
