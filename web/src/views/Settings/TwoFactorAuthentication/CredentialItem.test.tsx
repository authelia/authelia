import { fireEvent, render, screen } from "@testing-library/react";

import CredentialItem from "@views/Settings/TwoFactorAuthentication/CredentialItem";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@hooks/RelativeTimeString", () => ({
    useRelativeTime: () => "2 days ago",
}));

const baseProps = {
    created_at: new Date("2024-01-01"),
    description: "My Key",
    handleDelete: vi.fn(),
    id: "test-credential",
    qualifier: "(FIDO2)",
    tooltipDelete: "Delete this",
};

beforeEach(() => {
    baseProps.handleDelete.mockReset();
});

it("renders description and qualifier", () => {
    render(<CredentialItem {...baseProps} />);
    expect(screen.getByText("My Key")).toBeInTheDocument();
    expect(screen.getByText("(FIDO2)")).toBeInTheDocument();
});

it("renders added and never used timestamps", () => {
    render(<CredentialItem {...baseProps} />);
    expect(screen.getByText("Added 2 days ago")).toBeInTheDocument();
    expect(screen.getByText("Never used")).toBeInTheDocument();
});

it("renders last used when provided", () => {
    render(<CredentialItem {...baseProps} last_used_at={new Date("2024-06-01")} />);
    expect(screen.getByText("Last Used 2 days ago")).toBeInTheDocument();
});

it("calls handleDelete when delete button is clicked", () => {
    const { container } = render(<CredentialItem {...baseProps} />);
    const deleteButton = container.querySelector("#test-credential-delete") as HTMLElement;
    fireEvent.click(deleteButton);
    expect(baseProps.handleDelete).toHaveBeenCalledOnce();
});

it("renders information button when handleInformation is provided", () => {
    const handleInfo = vi.fn();
    const { container } = render(
        <CredentialItem {...baseProps} handleInformation={handleInfo} tooltipInformation="View info" />,
    );
    const infoButton = container.querySelector("#test-credential-information") as HTMLElement;
    fireEvent.click(infoButton);
    expect(handleInfo).toHaveBeenCalledOnce();
});

it("renders edit button when handleEdit is provided", () => {
    const handleEdit = vi.fn();
    const { container } = render(<CredentialItem {...baseProps} handleEdit={handleEdit} tooltipEdit="Edit this" />);
    const editButton = container.querySelector("#test-credential-edit") as HTMLElement;
    fireEvent.click(editButton);
    expect(handleEdit).toHaveBeenCalledOnce();
});

it("renders problem icon when problem flag is set", () => {
    const { container } = render(
        <CredentialItem
            {...baseProps}
            problem={true}
            handleInformation={vi.fn()}
            tooltipInformationProblem="There is a problem"
        />,
    );
    const infoButton = container.querySelector("#test-credential-information") as HTMLElement;
    expect(infoButton).toBeInTheDocument();
    expect(infoButton.querySelector("svg")).toBeInTheDocument();
});
