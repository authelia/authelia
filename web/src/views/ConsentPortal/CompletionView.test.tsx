import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";

import CompletionView from "@views/ConsentPortal/CompletionView";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

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

const renderWithRouter = (search: string) => {
    return render(
        <MemoryRouter initialEntries={[`/consent/completion${search}`]}>
            <CompletionView />
        </MemoryRouter>,
    );
};

it("renders accepted decision title", () => {
    renderWithRouter("?decision=accepted");

    expect(screen.getByTestId("login-layout")).toHaveAttribute("data-title", "Consent has been accepted and processed");
});

it("renders rejected decision title", () => {
    renderWithRouter("?decision=rejected");

    expect(screen.getByTestId("login-layout")).toHaveAttribute("data-title", "Consent has been rejected and processed");
});

it("renders error title and error details when error param is present", () => {
    renderWithRouter("?error=invalid_request&error_description=Bad+request&error_hint=Check+params");

    expect(screen.getByTestId("login-layout")).toHaveAttribute(
        "data-title",
        "An error occurred processing the request",
    );
    expect(screen.getByText(/invalid_request/)).toBeInTheDocument();
    expect(screen.getByText(/Bad request/)).toBeInTheDocument();
    expect(screen.getByText(/Check params/)).toBeInTheDocument();
});

it("renders home button", () => {
    renderWithRouter("?decision=accepted");

    expect(screen.getByTestId("home-button")).toBeInTheDocument();
});

it("marks an accepted decision as successful", () => {
    renderWithRouter("?decision=accepted");

    expect(screen.getByTestId("openid-completion-outcome")).toHaveAttribute("data-outcome", "accepted");
});

it("marks a rejected decision", () => {
    renderWithRouter("?decision=rejected");

    expect(screen.getByTestId("openid-completion-outcome")).toHaveAttribute("data-outcome", "rejected");
});

it("marks an errored request", () => {
    renderWithRouter("?error=invalid_request");

    expect(screen.getByTestId("openid-completion-outcome")).toHaveAttribute("data-outcome", "error");
});

it("labels each error detail", () => {
    renderWithRouter("?error=invalid_request&error_description=Bad+request&error_hint=Check+params");

    expect(screen.getByText("Error")).toBeInTheDocument();
    expect(screen.getByText("Description")).toBeInTheDocument();
    expect(screen.getByText("Hint")).toBeInTheDocument();
});

it("omits error details that were not supplied", () => {
    renderWithRouter("?error=invalid_request");

    expect(screen.queryByText("Description")).not.toBeInTheDocument();
    expect(screen.queryByText("Hint")).not.toBeInTheDocument();
    expect(screen.queryByText("Documentation")).not.toBeInTheDocument();
});

it("hides debug information behind a disclosure", () => {
    renderWithRouter("?error=invalid_request&error_debug=stack+trace+here");

    expect(screen.queryByText("stack trace here")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Debug Information" })).toBeInTheDocument();
});

it("reveals debug information when the disclosure is opened", () => {
    renderWithRouter("?error=invalid_request&error_debug=stack+trace+here");

    fireEvent.click(screen.getByRole("button", { name: "Debug Information" }));

    expect(screen.getByText("stack trace here")).toBeInTheDocument();
});

it("does not render a debug disclosure when there is nothing to debug", () => {
    renderWithRouter("?error=invalid_request");

    expect(screen.queryByRole("button", { name: "Debug Information" })).not.toBeInTheDocument();
});
