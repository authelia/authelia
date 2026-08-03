import { fireEvent, render, screen } from "@testing-library/react";

import PasskeyRememberMeDialog from "@views/LoginPortal/FirstFactor/PasskeyRememberMeDialog";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

it("renders nothing when closed", () => {
    render(<PasskeyRememberMeDialog open={false} onChoice={vi.fn()} />);
    expect(screen.queryByText("Remember me?")).not.toBeInTheDocument();
});

it("renders the prompt when open", () => {
    render(<PasskeyRememberMeDialog open={true} onChoice={vi.fn()} />);
    expect(screen.getByText("Remember me?")).toBeInTheDocument();
    expect(screen.getByText("Would you like to stay signed in on this device?")).toBeInTheDocument();
});

it("calls onChoice with true when yes is clicked", () => {
    const onChoice = vi.fn();
    render(<PasskeyRememberMeDialog open={true} onChoice={onChoice} />);
    fireEvent.click(screen.getByRole("button", { name: "Yes" }));
    expect(onChoice).toHaveBeenCalledWith(true);
});

it("calls onChoice with false when no is clicked", () => {
    const onChoice = vi.fn();
    render(<PasskeyRememberMeDialog open={true} onChoice={onChoice} />);
    fireEvent.click(screen.getByRole("button", { name: "No" }));
    expect(onChoice).toHaveBeenCalledWith(false);
});

it("calls onChoice with false when dismissed", () => {
    const onChoice = vi.fn();
    render(<PasskeyRememberMeDialog open={true} onChoice={onChoice} />);
    fireEvent.keyDown(screen.getByRole("dialog"), { key: "Escape" });
    expect(onChoice).toHaveBeenCalledWith(false);
});
