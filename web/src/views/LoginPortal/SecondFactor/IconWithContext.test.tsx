import { render, screen } from "@testing-library/react";

import IconWithContext from "@views/LoginPortal/SecondFactor/IconWithContext";

it("renders icon and children", () => {
    render(
        <IconWithContext icon={<span data-testid="test-icon" />}>
            <span data-testid="test-children">Content</span>
        </IconWithContext>,
    );
    expect(screen.getByTestId("test-icon")).toBeInTheDocument();
    expect(screen.getByTestId("test-children")).toHaveTextContent("Content");
});
