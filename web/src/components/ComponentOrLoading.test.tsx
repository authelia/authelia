import { render, screen } from "@testing-library/react";

import ComponentOrLoading from "@components/ComponentOrLoading";

vi.mock("@views/LoadingPage/LoadingPage", () => ({
    default: () => <div>Loading</div>,
}));

it("renders without crashing", () => {
    render(
        <ComponentOrLoading ready={false}>
            <div>Child</div>
        </ComponentOrLoading>,
    );
});

it("renders loading component and does not render children when not ready", () => {
    render(
        <ComponentOrLoading ready={false}>
            <div>Child</div>
        </ComponentOrLoading>,
    );

    expect(screen.getByText("Loading")).toBeInTheDocument();

    const wrapper = screen.getByText("Loading").parentElement;
    expect(wrapper).not.toHaveClass("hidden");

    expect(screen.queryByText("Child")).not.toBeInTheDocument();
});

it("renders children and hides loading component when ready", () => {
    render(
        <ComponentOrLoading ready={true}>
            <div>Child</div>
        </ComponentOrLoading>,
    );

    expect(screen.getByText("Child")).toBeInTheDocument();

    const wrapper = screen.getByText("Loading").parentElement;
    expect(wrapper).toHaveClass("hidden");

    expect(screen.getByText("Loading")).toBeInTheDocument();
});
