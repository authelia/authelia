import { fireEvent, render, screen } from "@testing-library/react";

import HomeButton from "@components/HomeButton";

const mockNavigate = vi.fn();

vi.mock("@hooks/RouterNavigate", () => ({
    useRouterNavigate: vi.fn(() => mockNavigate),
}));

vi.mock("@constants/Routes", () => ({
    IndexRoute: "/",
}));

it("renders without crashing", () => {
    render(<HomeButton />);
});

it("renders button with home text", () => {
    render(<HomeButton />);
    expect(screen.getByText("Home")).toBeInTheDocument();
});

it("navigates on click", () => {
    render(<HomeButton />);
    const button = screen.getByRole("button");
    fireEvent.click(button);
    expect(mockNavigate).toHaveBeenCalledWith("/", false, false, false);
});

it("renders as an outlined button rather than a bare ghost", () => {
    render(<HomeButton />);

    expect(screen.getByRole("button")).toHaveAttribute("data-variant", "outline");
});

it("renders a home icon alongside the label", () => {
    const { container } = render(<HomeButton />);

    expect(container.querySelector("svg")).toBeInTheDocument();
});
