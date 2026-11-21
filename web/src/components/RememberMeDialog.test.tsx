// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import RememberMeDialog from "@components/RememberMeDialog";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

it("renders nothing when closed", () => {
    render(<RememberMeDialog open={false} onCancel={vi.fn()} onChoice={vi.fn()} />);

    expect(screen.queryByText("Remember me")).not.toBeInTheDocument();
});

it("renders the prompt when open", () => {
    render(<RememberMeDialog open={true} onCancel={vi.fn()} onChoice={vi.fn()} />);

    expect(document.getElementById("remember-me-dialog")).toBeInTheDocument();
    expect(screen.getByText("Remember me")).toBeInTheDocument();
    expect(screen.getByText("Would you like to stay signed in on this device?")).toBeInTheDocument();
});

it("calls onChoice with true when yes is clicked", () => {
    const onChoice = vi.fn();

    render(<RememberMeDialog open={true} onCancel={vi.fn()} onChoice={onChoice} />);

    fireEvent.click(screen.getByRole("button", { name: "Yes" }));

    expect(onChoice).toHaveBeenCalledWith(true);
});

it("calls onChoice with false when no is clicked", () => {
    const onChoice = vi.fn();

    render(<RememberMeDialog open={true} onCancel={vi.fn()} onChoice={onChoice} />);

    fireEvent.click(screen.getByRole("button", { name: "No" }));

    expect(onChoice).toHaveBeenCalledWith(false);
});

it("calls onCancel rather than onChoice when dismissed", async () => {
    const onCancel = vi.fn();
    const onChoice = vi.fn();

    render(<RememberMeDialog open={true} onCancel={onCancel} onChoice={onChoice} />);

    fireEvent.keyDown(screen.getByRole("dialog"), { key: "Escape" });

    await waitFor(() => expect(onCancel).toHaveBeenCalled());
    expect(onChoice).not.toHaveBeenCalled();
});
