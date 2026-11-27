// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { fireEvent, render, screen } from "@testing-library/react";

import DecisionFormClient from "@views/ConsentPortal/OpenIDConnect/DecisionFormClient";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

it("renders the client description as the name", () => {
    render(<DecisionFormClient client_id={"test-client"} client_description={"Test Client"} />);

    expect(screen.getByText("Test Client")).toBeInTheDocument();
});

it("renders the client id alongside the description", () => {
    render(<DecisionFormClient client_id={"test-client"} client_description={"Test Client"} />);

    expect(screen.getByText("test-client")).toBeInTheDocument();
});

it("renders the client id as the name when the description is empty", () => {
    render(<DecisionFormClient client_id={"test-client"} client_description={""} />);

    expect(screen.getByTestId("openid-consent-client-name")).toHaveTextContent("test-client");
});

it("does not repeat the client id when the description is empty", () => {
    render(<DecisionFormClient client_id={"test-client"} client_description={""} />);

    expect(screen.getAllByText("test-client")).toHaveLength(1);
});

it("does not render a logo when no logo uri is provided", () => {
    render(<DecisionFormClient client_id={"test-client"} client_description={"Test Client"} />);

    expect(screen.queryByRole("img")).toBeNull();
});

it("renders the logo with the client description as the alt text", () => {
    render(
        <DecisionFormClient
            client_id={"test-client"}
            client_description={"Test Client"}
            client_logo_uri={"https://example.com/logo.png"}
        />,
    );

    const logo = screen.getByRole("img");

    expect(logo).toHaveAttribute("src", "https://example.com/logo.png");
    expect(logo).toHaveAttribute("alt", "Test Client");
});

it("renders the logo with the client id as the alt text when the description is empty", () => {
    render(
        <DecisionFormClient
            client_id={"test-client"}
            client_description={""}
            client_logo_uri={"https://example.com/logo.png"}
        />,
    );

    expect(screen.getByRole("img")).toHaveAttribute("alt", "test-client");
});

it("falls back to the placeholder icon when the logo fails to load", () => {
    const { container } = render(
        <DecisionFormClient
            client_id={"test-client"}
            client_description={"Test Client"}
            client_logo_uri={"https://example.com/logo.png"}
        />,
    );

    fireEvent.error(screen.getByRole("img"));

    expect(screen.queryByRole("img")).toBeNull();
    expect(container.querySelector("svg")).not.toBeNull();
});
