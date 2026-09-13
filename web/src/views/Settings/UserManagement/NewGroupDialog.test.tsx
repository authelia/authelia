import { act, fireEvent, render, screen } from "@testing-library/react";

import { postNewGroup } from "@services/GroupManagement";
import NewGroupDialog from "@views/Settings/UserManagement/NewGroupDialog";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string, opts?: any) => (opts?.item ? `${key}:${opts.item}` : key) }),
}));

const mockCreateError = vi.fn();
const mockCreateSuccess = vi.fn();

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: mockCreateError,
        createSuccessNotification: mockCreateSuccess,
    }),
}));

vi.mock("@services/GroupManagement", () => ({
    postNewGroup: vi.fn(),
}));

const onClose = vi.fn();

const name = () => document.getElementById("new-group-name") as HTMLInputElement;
const submit = () => document.getElementById("new-group-submit") as HTMLButtonElement;
const cancel = () => document.getElementById("new-group-cancel") as HTMLButtonElement;

beforeEach(() => {
    vi.spyOn(console, "log").mockImplementation(() => {});
    vi.mocked(postNewGroup).mockReset();
    mockCreateError.mockReset();
    mockCreateSuccess.mockReset();
    onClose.mockReset();
});

afterEach(() => {
    vi.restoreAllMocks();
});

it("renders the title and the group name field", () => {
    render(<NewGroupDialog open={true} onClose={onClose} />);

    expect(screen.getByText("New {{item}}:Group")).toBeInTheDocument();
    expect(name()).toBeInTheDocument();
    expect(name()).toBeRequired();
});

it("keeps save disabled until the name is changed", () => {
    render(<NewGroupDialog open={true} onClose={onClose} />);

    expect(submit()).toBeDisabled();

    fireEvent.change(name(), { target: { value: "dev" } });
    expect(submit()).toBeEnabled();

    fireEvent.change(name(), { target: { value: "" } });
    expect(submit()).toBeDisabled();
});

it("creates the group, notifies and closes on success", async () => {
    vi.mocked(postNewGroup).mockResolvedValue(undefined);

    render(<NewGroupDialog open={true} onClose={onClose} />);

    fireEvent.change(name(), { target: { value: "dev" } });

    await act(async () => {
        fireEvent.click(submit());
    });

    expect(postNewGroup).toHaveBeenCalledWith({ name: "dev" });
    expect(mockCreateSuccess).toHaveBeenCalledWith("Group created successfully.");
    expect(onClose).toHaveBeenCalledOnce();
});

it("notifies an error and keeps the dialog open on failure", async () => {
    vi.mocked(postNewGroup).mockRejectedValue(new Error("boom"));

    render(<NewGroupDialog open={true} onClose={onClose} />);

    fireEvent.change(name(), { target: { value: "dev" } });

    await act(async () => {
        fireEvent.click(submit());
    });

    expect(mockCreateError).toHaveBeenCalledWith("Error creating group");
    expect(onClose).not.toHaveBeenCalled();
    expect(name().value).toBe("dev");
});

it("closes without creating when cancel is clicked", () => {
    render(<NewGroupDialog open={true} onClose={onClose} />);

    fireEvent.change(name(), { target: { value: "dev" } });
    fireEvent.click(cancel());

    expect(postNewGroup).not.toHaveBeenCalled();
    expect(onClose).toHaveBeenCalledOnce();
});

it("resets the form when reopened after closing", () => {
    const { rerender } = render(<NewGroupDialog open={true} onClose={onClose} />);

    fireEvent.change(name(), { target: { value: "dev" } });

    rerender(<NewGroupDialog open={false} onClose={onClose} />);
    rerender(<NewGroupDialog open={true} onClose={onClose} />);

    expect(name().value).toBe("");
    expect(submit()).toBeDisabled();
});
