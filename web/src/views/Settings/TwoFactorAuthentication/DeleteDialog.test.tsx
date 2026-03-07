import { fireEvent, render, screen } from "@testing-library/react";

import DeleteDialog from "@views/Settings/TwoFactorAuthentication/DeleteDialog";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

it("renders the dialog with title and text when open", () => {
    render(
        <DeleteDialog open={true} title="Delete Item" text="Are you sure?" onConfirm={vi.fn()} onCancel={vi.fn()} />,
    );
    expect(screen.getByText("Delete Item")).toBeInTheDocument();
    expect(screen.getByText("Are you sure?")).toBeInTheDocument();
});

it("calls onCancel when cancel button is clicked", () => {
    const onCancel = vi.fn();
    render(<DeleteDialog open={true} title="Delete" text="Sure?" onConfirm={vi.fn()} onCancel={onCancel} />);
    fireEvent.click(screen.getByText("Cancel"));
    expect(onCancel).toHaveBeenCalledOnce();
});

it("calls onConfirm when remove button is clicked", () => {
    const onConfirm = vi.fn();
    render(<DeleteDialog open={true} title="Delete" text="Sure?" onConfirm={onConfirm} onCancel={vi.fn()} />);
    fireEvent.click(screen.getByText("Remove"));
    expect(onConfirm).toHaveBeenCalledOnce();
});

it("does not render content when closed", () => {
    render(<DeleteDialog open={false} title="Delete" text="Sure?" onConfirm={vi.fn()} onCancel={vi.fn()} />);
    expect(screen.queryByText("Delete")).not.toBeInTheDocument();
});
