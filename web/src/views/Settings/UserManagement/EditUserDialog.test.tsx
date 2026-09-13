import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";

import { UserDetailsExtended } from "@models/UserManagement";
import { patchChangeUser } from "@services/UserManagement";
import EditUserDialog from "@views/Settings/UserManagement/EditUserDialog";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({
        t: (key: string, opts?: any) => opts?.defaultValue ?? (opts?.item ? `${key}:${opts.item}` : key),
    }),
}));

const mockCreateError = vi.fn();
const mockCreateSuccess = vi.fn();

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: mockCreateError,
        createSuccessNotification: mockCreateSuccess,
    }),
}));

const mocks = vi.hoisted(() => ({
    groups: undefined as string[] | undefined,
    metadata: undefined as any,
    refetchGroups: vi.fn(),
    refetchMetadata: vi.fn(),
}));

vi.mock("@hooks/UserManagement", () => ({
    useUserManagementAttributeMetadataGET: () => [mocks.metadata, mocks.refetchMetadata, false, undefined],
}));

vi.mock("@hooks/GroupManagement", () => ({
    useAllGroupsGET: () => [mocks.groups, mocks.refetchGroups, false, undefined],
}));

vi.mock("@services/UserManagement", () => ({
    patchChangeUser: vi.fn(),
}));

const metadata = {
    required_attributes: ["username"],
    supported_attributes: {
        custom_field: { type: "text" },
        display_name: { type: "text" },
        groups: { multiple: true, type: "groups" },
        last_logged_in: { type: "text" },
        locality: { type: "text" },
        mail: { multiple: true, type: "email" },
        password: { type: "password" },
        username: { type: "text" },
    },
};

const user: UserDetailsExtended = {
    address: { locality: "Paris" },
    display_name: "John",
    extra: { custom_field: "a" },
    groups: ["dev"],
    mail: ["john@example.com"],
    username: "john",
};

const onClose = vi.fn();

const byId = (id: string) => document.getElementById(id) as HTMLInputElement;
const submit = () => byId("edit-user-submit") as unknown as HTMLButtonElement;
const cancel = () => byId("edit-user-cancel") as unknown as HTMLButtonElement;

beforeEach(() => {
    vi.spyOn(console, "log").mockImplementation(() => {});
    vi.mocked(patchChangeUser).mockReset();
    mockCreateError.mockReset();
    mockCreateSuccess.mockReset();
    onClose.mockReset();
    mocks.groups = ["admins", "dev"];
    mocks.metadata = metadata;
    mocks.refetchGroups.mockReset();
    mocks.refetchMetadata.mockReset();
});

afterEach(() => {
    vi.restoreAllMocks();
});

it("fetches the attribute metadata and the groups when opened", () => {
    render(<EditUserDialog open={true} user={user} onClose={onClose} />);

    expect(mocks.refetchMetadata).toHaveBeenCalled();
    expect(mocks.refetchGroups).toHaveBeenCalled();
});

it("renders the title with the username and a disabled username field", () => {
    render(<EditUserDialog open={true} user={user} onClose={onClose} />);

    expect(screen.getByText(/^Edit .* john$/)).toBeInTheDocument();
    expect(byId("edit-user-username")).toBeDisabled();
    expect(byId("edit-user-username").value).toBe("john");
    expect(byId("edit-user-password")).toBeNull();
    expect(byId("edit-user-last_logged_in")).toBeNull();
});

it("pre-fills the editable fields from the user", () => {
    render(<EditUserDialog open={true} user={user} onClose={onClose} />);

    fireEvent.click(screen.getByText("Show Additional Fields"));

    expect(byId("edit-user-display_name").value).toBe("John");
    expect(byId("edit-user-address-locality").value).toBe("Paris");
});

it("keeps save disabled until a field is changed", () => {
    render(<EditUserDialog open={true} user={user} onClose={onClose} />);

    expect(submit()).toBeDisabled();

    fireEvent.click(screen.getByText("Show Additional Fields"));
    fireEvent.change(byId("edit-user-display_name"), { target: { value: "Johnny" } });

    expect(submit()).toBeEnabled();
});

it("patches only the changed top-level fields with a matching update mask", async () => {
    vi.mocked(patchChangeUser).mockResolvedValue(undefined);

    render(<EditUserDialog open={true} user={user} onClose={onClose} />);

    fireEvent.click(screen.getByText("Show Additional Fields"));
    fireEvent.change(byId("edit-user-display_name"), { target: { value: "Johnny" } });

    await act(async () => {
        fireEvent.click(submit());
    });

    await waitFor(() => expect(patchChangeUser).toHaveBeenCalledOnce());
    expect(patchChangeUser).toHaveBeenCalledWith("john", { display_name: "Johnny" }, ["display_name"]);
    expect(mockCreateSuccess).toHaveBeenCalledWith("User modified successfully.");
    expect(onClose).toHaveBeenCalledOnce();
});

it("prefixes address fields in the update mask", async () => {
    vi.mocked(patchChangeUser).mockResolvedValue(undefined);

    render(<EditUserDialog open={true} user={user} onClose={onClose} />);

    fireEvent.click(screen.getByText("Show Additional Fields"));
    fireEvent.change(byId("edit-user-address-locality"), { target: { value: "Lyon" } });

    await act(async () => {
        fireEvent.click(submit());
    });

    await waitFor(() => expect(patchChangeUser).toHaveBeenCalledOnce());

    const [username, body, mask] = vi.mocked(patchChangeUser).mock.calls[0];
    expect(username).toBe("john");
    expect(mask).toEqual(["address.locality"]);
    expect(body).toEqual({ address: { locality: "Lyon" } });
});

it("prefixes extra fields in the update mask and nests them under extra", async () => {
    vi.mocked(patchChangeUser).mockResolvedValue(undefined);

    render(<EditUserDialog open={true} user={user} onClose={onClose} />);

    fireEvent.click(screen.getByText("Show Additional Fields"));
    fireEvent.change(byId("edit-user-extra-custom_field"), { target: { value: "b" } });

    await act(async () => {
        fireEvent.click(submit());
    });

    await waitFor(() => expect(patchChangeUser).toHaveBeenCalledOnce());
    expect(patchChangeUser).toHaveBeenCalledWith("john", { extra: { custom_field: "b" } }, ["extra.custom_field"]);
});

it("notifies an error and keeps the dialog open when the patch fails", async () => {
    vi.mocked(patchChangeUser).mockRejectedValue(new Error("boom"));

    render(<EditUserDialog open={true} user={user} onClose={onClose} />);

    fireEvent.click(screen.getByText("Show Additional Fields"));
    fireEvent.change(byId("edit-user-display_name"), { target: { value: "Johnny" } });

    await act(async () => {
        fireEvent.click(submit());
    });

    await waitFor(() => expect(mockCreateError).toHaveBeenCalledWith("Error modifying user"));
    expect(onClose).not.toHaveBeenCalled();
    expect(byId("edit-user-display_name").value).toBe("Johnny");
});

it("closes immediately on cancel when nothing changed", () => {
    render(<EditUserDialog open={true} user={user} onClose={onClose} />);

    fireEvent.click(cancel());

    expect(screen.queryByText("Unsaved Changes")).not.toBeInTheDocument();
    expect(onClose).toHaveBeenCalledOnce();
});

it("asks for confirmation on cancel when there are unsaved changes", () => {
    render(<EditUserDialog open={true} user={user} onClose={onClose} />);

    fireEvent.click(screen.getByText("Show Additional Fields"));
    fireEvent.change(byId("edit-user-display_name"), { target: { value: "Johnny" } });
    fireEvent.click(cancel());

    expect(screen.getByText("Unsaved Changes")).toBeInTheDocument();
    expect(onClose).not.toHaveBeenCalled();

    fireEvent.click(byId("verify-exit-cancel"));

    expect(onClose).not.toHaveBeenCalled();
    expect(byId("edit-user-display_name").value).toBe("Johnny");

    fireEvent.click(cancel());
    fireEvent.click(byId("verify-exit-confirm"));

    expect(onClose).toHaveBeenCalledOnce();
    expect(patchChangeUser).not.toHaveBeenCalled();
});

it("renders nothing when closed", () => {
    render(<EditUserDialog open={false} user={user} onClose={onClose} />);

    expect(byId("edit-user-username")).toBeNull();
    expect(mocks.refetchMetadata).not.toHaveBeenCalled();
});
