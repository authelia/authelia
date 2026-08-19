import { render, screen } from "@testing-library/react";

import BaseLoadingPage from "@views/LoadingPage/BaseLoadingPage";

it("renders the loading message", () => {
    render(<BaseLoadingPage message="Please wait" />);
    expect(screen.getByText("Please wait...")).toBeInTheDocument();
});

it("renders the loading bars", () => {
    const { container } = render(<BaseLoadingPage message="Loading" />);
    expect(container.querySelectorAll(".animate-scale-loader")).toHaveLength(5);
});

it("does not inject a stylesheet to animate them", () => {
    const before = document.head.querySelectorAll("style").length;

    render(<BaseLoadingPage message="Loading" />);

    expect(document.head.querySelectorAll("style")).toHaveLength(before);
});
