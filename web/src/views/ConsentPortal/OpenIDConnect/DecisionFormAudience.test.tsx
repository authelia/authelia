import { render, screen } from "@testing-library/react";

import DecisionFormAudience from "@views/ConsentPortal/OpenIDConnect/DecisionFormAudience";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

it("renders an item for each audience", () => {
    render(<DecisionFormAudience audience={["https://api.example.com", "https://files.example.com"]} />);

    expect(screen.getByText("https://api.example.com")).toBeInTheDocument();
    expect(screen.getByText("https://files.example.com")).toBeInTheDocument();
    expect(screen.getAllByRole("listitem")).toHaveLength(2);
});

it("assigns a predictable id to each audience item", () => {
    render(<DecisionFormAudience audience={["https://api.example.com"]} />);

    expect(screen.getByRole("listitem")).toHaveAttribute("id", "audience-https://api.example.com");
});

it("renders the section heading", () => {
    render(<DecisionFormAudience audience={["https://api.example.com"]} />);

    expect(screen.getByRole("heading", { name: "Token Audience" })).toBeInTheDocument();
});

it("renders nothing when the audience is empty", () => {
    const { container } = render(<DecisionFormAudience audience={[]} />);

    expect(container).toBeEmptyDOMElement();
});

it("renders nothing when the audience is null", () => {
    const { container } = render(<DecisionFormAudience audience={null} />);

    expect(container).toBeEmptyDOMElement();
});

it("renders nothing when the audience is undefined", () => {
    const { container } = render(<DecisionFormAudience audience={undefined} />);

    expect(container).toBeEmptyDOMElement();
});
