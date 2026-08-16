import { render, screen } from "@testing-library/react";

import BaseLoadingPage from "@views/LoadingPage/BaseLoadingPage";

vi.mock("react-spinners", () => ({
    ScaleLoader: () => <div data-testid="scale-loader" />,
}));

it("renders the loading message", () => {
    render(<BaseLoadingPage message="Please wait" />);
    expect(screen.getByText("Please wait...")).toBeInTheDocument();
});

it("renders the scale loader", () => {
    render(<BaseLoadingPage message="Loading" />);
    expect(screen.getByTestId("scale-loader")).toBeInTheDocument();
});
