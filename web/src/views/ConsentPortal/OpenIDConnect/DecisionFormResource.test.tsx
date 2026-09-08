import { render, screen } from "@testing-library/react";

import DecisionFormResource from "@views/ConsentPortal/OpenIDConnect/DecisionFormResource";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

it("renders an item for each resource", () => {
    render(<DecisionFormResource resource={["https://api.example.com", "https://files.example.com"]} />);

    expect(screen.getByText("https://api.example.com")).toBeInTheDocument();
    expect(screen.getByText("https://files.example.com")).toBeInTheDocument();
    expect(screen.getAllByRole("listitem")).toHaveLength(2);
});

it("assigns a predictable id to each resource item", () => {
    render(<DecisionFormResource resource={["https://api.example.com"]} />);

    expect(screen.getByRole("listitem")).toHaveAttribute("id", "resource-https://api.example.com");
});

it("renders the section heading", () => {
    render(<DecisionFormResource resource={["https://api.example.com"]} />);

    expect(screen.getByRole("heading", { name: "Requested Resources" })).toBeInTheDocument();
});

it("renders nothing when the resource is empty", () => {
    const { container } = render(<DecisionFormResource resource={[]} />);

    expect(container).toBeEmptyDOMElement();
});

it("renders nothing when the resource is null", () => {
    const { container } = render(<DecisionFormResource resource={null} />);

    expect(container).toBeEmptyDOMElement();
});

it("renders nothing when the resource is undefined", () => {
    const { container } = render(<DecisionFormResource resource={undefined} />);

    expect(container).toBeEmptyDOMElement();
});
