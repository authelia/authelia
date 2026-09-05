import { render, screen } from "@testing-library/react";
import i18n from "i18next";
import { MemoryRouter } from "react-router";

import CompletionView from "@views/ConsentPortal/CompletionView";

vi.mock("@components/HomeButton", () => ({
    default: () => <button data-testid="home-button">Home</button>,
}));

vi.mock("@layouts/LoginLayout", () => ({
    default: (props: any) => (
        <div data-testid="login-layout" data-title={props.title}>
            {props.children}
        </div>
    ),
}));

beforeAll(async () => {
    i18n.addResourceBundle("en", "consent", {
        "An error occurred processing the request": "An error occurred processing the request",
        "Debug Information": "Debug Information",
        Description: "Description",
        Documentation: "Documentation",
        Error: "Error",
        Hint: "Hint",
    });

    await i18n.changeLanguage("en");
});

const renderWithRouter = (search: string) =>
    render(
        <MemoryRouter initialEntries={[`/consent/completion${search}`]}>
            <CompletionView />
        </MemoryRouter>,
    );

it("keeps the scheme of the documentation url", () => {
    renderWithRouter("?error=invalid_request&error_uri=https%3A%2F%2Fwww.authelia.com%2Fdocs");

    expect(screen.getByText("https://www.authelia.com/docs")).toBeInTheDocument();
});

it("does not attempt to translate the error code", () => {
    renderWithRouter("?error=access_denied");

    expect(screen.getByText("access_denied")).toBeInTheDocument();
});

it("keeps a colon inside an error description", () => {
    renderWithRouter("?error=invalid_client&error_description=Client%20authentication%20failed%3A%20bad%20secret");

    expect(screen.getByText("Client authentication failed: bad secret")).toBeInTheDocument();
});
