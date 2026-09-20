import { render, screen } from "@testing-library/react";

import { Card, CardAction, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@components/UI/Card";

it("renders a complete card", () => {
    const { container } = render(
        <Card className="card">
            <CardHeader className="header">
                <CardTitle>Title</CardTitle>
                <CardDescription>Description</CardDescription>
                <CardAction>Action</CardAction>
            </CardHeader>
            <CardContent className="content">Body</CardContent>
            <CardFooter className="footer">Footer</CardFooter>
        </Card>,
    );

    expect(container.querySelector('[data-slot="card"]')).toHaveClass("card");
    expect(container.querySelector('[data-slot="card-header"]')).toHaveClass("header");
    expect(container.querySelector('[data-slot="card-content"]')).toHaveClass("content");
    expect(container.querySelector('[data-slot="card-footer"]')).toHaveClass("footer");
    expect(screen.getByText("Title")).toHaveAttribute("data-slot", "card-title");
    expect(screen.getByText("Description")).toHaveAttribute("data-slot", "card-description");
    expect(screen.getByText("Action")).toHaveAttribute("data-slot", "card-action");
});

it("renders a bare card", () => {
    const { container } = render(<Card>Bare</Card>);
    expect(container.querySelector('[data-slot="card"]')).toHaveTextContent("Bare");
});

it("forwards arbitrary props", () => {
    const { container } = render(<Card id="my-card" data-loading="true" />);
    const card = container.querySelector('[data-slot="card"]');
    expect(card).toHaveAttribute("id", "my-card");
    expect(card).toHaveAttribute("data-loading", "true");
});
