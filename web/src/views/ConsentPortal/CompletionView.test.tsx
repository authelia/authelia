import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";

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
