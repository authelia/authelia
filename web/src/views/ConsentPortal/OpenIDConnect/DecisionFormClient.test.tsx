import { render, screen } from "@testing-library/react";

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
