import { act, fireEvent, render, screen } from "@testing-library/react";

import GroupManagementView from "@views/Settings/UserManagement/GroupManagementView";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({
        t: (key: string, opts?: any) => (opts ? `${key}:${Object.values(opts).join(",")}` : key),
    }),
}));

const mockCreateError = vi.fn();

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: mockCreateError,
        createSuccessNotification: vi.fn(),
    }),
}));

const mocks = vi.hoisted(() => ({
    error: undefined as Error | undefined,
    fetchGroups: vi.fn(),
    groups: undefined as string[] | undefined,
}));

vi.mock("@hooks/GroupManagement", () => ({
    useAllGroupsGET: () => [mocks.groups, mocks.fetchGroups, false, mocks.error],
}));

vi.mock("@views/Settings/UserManagement/NewGroupDialog", () => ({
    default: (props: any) => (
        <div data-testid="new-group-dialog" data-open={String(props.open)}>
            <button data-testid="new-group-close" onClick={() => props.onClose()} />
        </div>
    ),
}));

vi.mock("@views/Settings/UserManagement/VerifyDeleteGroupDialog", () => ({
    default: (props: any) => (
        <div data-testid="verify-delete-group-dialog" data-open={String(props.open)} data-group={props.groupName}>
            <button data-testid="verify-delete-group-close" onClick={() => props.onCancel()} />
        </div>
    ),
}));

const byId = (id: string) => document.getElementById(id) as HTMLElement;
const dialog = (testId: string) => screen.getByTestId(testId);

const renderView = async () => {
    await act(async () => {
        render(<GroupManagementView />);
    });
};

beforeEach(() => {
    mockCreateError.mockReset();
    mocks.error = undefined;
    mocks.fetchGroups.mockReset();
    mocks.groups = ["dev", "admins"];
});

it("fetches the groups on mount and renders the heading", async () => {
    await renderView();

    expect(mocks.fetchGroups).toHaveBeenCalledOnce();
    expect(screen.getByText("Group Management")).toBeInTheDocument();
    expect(byId("group-management-add")).toHaveTextContent("Add a {{item}}:group");
});

it("notifies when the groups cannot be fetched", async () => {
    mocks.groups = undefined;
    mocks.error = new Error("boom");

    await renderView();

    expect(mockCreateError).toHaveBeenCalledWith("There was an issue retrieving group info");
});

it("renders one row per group sorted by name", async () => {
    await renderView();

    const admins = screen.getByText("admins");
    const dev = screen.getByText("dev");

    expect(admins).toBeInTheDocument();
    expect(dev).toBeInTheDocument();
    expect(admins.compareDocumentPosition(dev) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
});

it("renders no rows while the groups are not loaded", async () => {
    mocks.groups = undefined;

    await renderView();

    expect(screen.queryByText("dev")).not.toBeInTheDocument();
    expect(mockCreateError).not.toHaveBeenCalled();
});

it("opens the new group dialog and refetches when it closes", async () => {
    await renderView();

    expect(dialog("new-group-dialog")).toHaveAttribute("data-open", "false");

    fireEvent.click(byId("group-management-add"));
    expect(dialog("new-group-dialog")).toHaveAttribute("data-open", "true");

    fireEvent.click(screen.getByTestId("new-group-close"));
    expect(dialog("new-group-dialog")).toHaveAttribute("data-open", "false");
    expect(mocks.fetchGroups).toHaveBeenCalledTimes(2);
});

it("opens the delete confirmation for the row's group and refetches when it closes", async () => {
    await renderView();

    fireEvent.click(byId("group-row-dev-delete"));

    expect(dialog("verify-delete-group-dialog")).toHaveAttribute("data-open", "true");
    expect(dialog("verify-delete-group-dialog")).toHaveAttribute("data-group", "dev");

    fireEvent.click(screen.getByTestId("verify-delete-group-close"));

    expect(dialog("verify-delete-group-dialog")).toHaveAttribute("data-open", "false");
    expect(mocks.fetchGroups).toHaveBeenCalledTimes(2);
});
