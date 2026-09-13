import { act, fireEvent, render, screen } from "@testing-library/react";

import { deleteGroup } from "@services/GroupManagement";
import VerifyDeleteGroupDialog from "@views/Settings/UserManagement/VerifyDeleteGroupDialog";

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
    deleteGroup: vi.fn(),
}));

const onCancel = vi.fn();

const input = () => document.getElementById("verify-delete-group-input") as HTMLInputElement;
const confirmButton = () => document.getElementById("verify-delete-group-confirm") as HTMLButtonElement;

beforeEach(() => {
    vi.spyOn(console, "log").mockImplementation(() => {});
    vi.mocked(deleteGroup).mockReset();
    mockCreateError.mockReset();
    mockCreateSuccess.mockReset();
    onCancel.mockReset();
});

afterEach(() => {
    vi.restoreAllMocks();
});

it("renders the title and the group name in the message", () => {
    render(<VerifyDeleteGroupDialog open={true} groupName="dev" onCancel={onCancel} />);

    expect(screen.getAllByText("Delete Group").length).toBeGreaterThanOrEqual(1);
    expect(
        screen.getByText("You are about to delete group {{item}}, enter the group name to continue.:dev"),
    ).toBeInTheDocument();
});

it("keeps the delete button disabled until the exact group name is typed", () => {
    render(<VerifyDeleteGroupDialog open={true} groupName="dev" onCancel={onCancel} />);

    expect(confirmButton()).toBeDisabled();

    fireEvent.change(input(), { target: { value: "de" } });
    expect(confirmButton()).toBeDisabled();

    fireEvent.change(input(), { target: { value: "dev" } });
    expect(confirmButton()).toBeEnabled();
});

it("deletes the group, notifies and closes on success", async () => {
    vi.mocked(deleteGroup).mockResolvedValue(undefined);

    render(<VerifyDeleteGroupDialog open={true} groupName="dev" onCancel={onCancel} />);

    fireEvent.change(input(), { target: { value: "dev" } });

    await act(async () => {
        fireEvent.click(confirmButton());
    });

    expect(deleteGroup).toHaveBeenCalledWith("dev");
    expect(mockCreateSuccess).toHaveBeenCalledWith("Group deleted successfully.");
    expect(onCancel).toHaveBeenCalledOnce();
});

it("notifies an error and still closes when deletion fails", async () => {
    vi.mocked(deleteGroup).mockRejectedValue(new Error("boom"));

    render(<VerifyDeleteGroupDialog open={true} groupName="dev" onCancel={onCancel} />);

    fireEvent.change(input(), { target: { value: "dev" } });

    await act(async () => {
        fireEvent.click(confirmButton());
    });

    expect(mockCreateError).toHaveBeenCalledWith("Error deleting group.");
    expect(onCancel).toHaveBeenCalledOnce();
});

it("closes without deleting when cancel is clicked", () => {
    render(<VerifyDeleteGroupDialog open={true} groupName="dev" onCancel={onCancel} />);

    fireEvent.click(document.getElementById("verify-delete-group-cancel")!);

    expect(deleteGroup).not.toHaveBeenCalled();
    expect(onCancel).toHaveBeenCalledOnce();
});
