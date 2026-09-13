import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";

import { postNewUser } from "@services/UserManagement";
import NewUserDialog from "@views/Settings/UserManagement/NewUserDialog";

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
    groupsError: undefined as Error | undefined,
    groupsLoading: false,
    metadata: undefined as any,
    metadataError: undefined as Error | undefined,
    metadataLoading: false,
    refetchGroups: vi.fn(),
    refetchMetadata: vi.fn(),
}));

vi.mock("@hooks/UserManagement", () => ({
    useUserManagementAttributeMetadataGET: () => [
        mocks.metadata,
        mocks.refetchMetadata,
        mocks.metadataLoading,
        mocks.metadataError,
    ],
}));

vi.mock("@hooks/GroupManagement", () => ({
    useAllGroupsGET: () => [mocks.groups, mocks.refetchGroups, mocks.groupsLoading, mocks.groupsError],
}));

vi.mock("@services/UserManagement", () => ({
    postNewUser: vi.fn(),
}));

vi.mock("@utils/GeneratePassword", () => ({
    generateRandomPassword: vi.fn(() => "Gen3rated!Pass"),
}));

const ldapMetadata = {
    required_attributes: ["username", "password"],
    supported_attributes: {
        birthdate: { type: "date" },
        custom_field: { type: "text" },
        display_name: { type: "text" },
        groups: { multiple: true, type: "groups" },
        last_logged_in: { type: "text" },
        mail: { multiple: true, type: "email" },
        password: { type: "password" },
        username: { type: "text" },
    },
};

const onClose = vi.fn();

const byId = (id: string) => document.getElementById(id) as HTMLInputElement;
const submit = () => byId("new-user-submit") as unknown as HTMLButtonElement;

beforeEach(() => {
    vi.spyOn(console, "log").mockImplementation(() => {});
    vi.mocked(postNewUser).mockReset();
    mockCreateError.mockReset();
    mockCreateSuccess.mockReset();
    onClose.mockReset();
    mocks.groups = ["admins", "dev"];
    mocks.groupsError = undefined;
    mocks.groupsLoading = false;
    mocks.metadata = ldapMetadata;
    mocks.metadataError = undefined;
    mocks.metadataLoading = false;
    mocks.refetchGroups.mockReset();
    mocks.refetchMetadata.mockReset();
});

afterEach(() => {
    vi.restoreAllMocks();
});

it("fetches the attribute metadata and the groups when opened", () => {
    render(<NewUserDialog open={true} onClose={onClose} />);

    expect(mocks.refetchMetadata).toHaveBeenCalled();
    expect(mocks.refetchGroups).toHaveBeenCalled();
});

it("does not fetch groups when the groups attribute is a plain text attribute", () => {
    mocks.metadata = {
        ...ldapMetadata,
        supported_attributes: { ...ldapMetadata.supported_attributes, groups: { multiple: true, type: "text" } },
    };

    render(<NewUserDialog open={true} onClose={onClose} />);

    expect(mocks.refetchMetadata).toHaveBeenCalled();
    expect(mocks.refetchGroups).not.toHaveBeenCalled();
});

it("does not fetch anything while closed", () => {
    render(<NewUserDialog open={false} onClose={onClose} />);

    expect(mocks.refetchMetadata).not.toHaveBeenCalled();
    expect(mocks.refetchGroups).not.toHaveBeenCalled();
    expect(screen.queryByText("New {{item}}:User")).not.toBeInTheDocument();
});

it("renders the basic fields and hides the additional ones behind a toggle", () => {
    render(<NewUserDialog open={true} onClose={onClose} />);

    expect(screen.getByText("New {{item}}:User")).toBeInTheDocument();
    expect(byId("new-user-username")).toBeInTheDocument();
    expect(byId("new-user-mail")).toBeInTheDocument();
    expect(byId("new-user-password")).toBeInTheDocument();
    expect(byId("new-user-groups")).toBeInTheDocument();
    expect(byId("new-user-display_name")).toBeNull();
    expect(byId("new-user-birthdate")).toBeNull();
    expect(byId("new-user-custom_field")).toBeNull();
    expect(byId("new-user-last_logged_in")).toBeNull();

    fireEvent.click(screen.getByText("Show Additional Fields"));

    expect(byId("new-user-display_name")).toBeInTheDocument();
    expect(byId("new-user-birthdate")).toHaveAttribute("type", "date");
    expect(byId("new-user-custom_field")).toBeInTheDocument();
    expect(byId("new-user-last_logged_in")).toBeNull();
    expect(screen.getByText("Hide Additional Fields")).toBeInTheDocument();

    fireEvent.click(screen.getByText("Hide Additional Fields"));

    expect(byId("new-user-display_name")).toBeNull();
});

it("keeps save disabled until the form is dirty", () => {
    render(<NewUserDialog open={true} onClose={onClose} />);

    expect(submit()).toBeDisabled();

    fireEvent.change(byId("new-user-username"), { target: { value: "jane" } });

    expect(submit()).toBeEnabled();
});

it("fills the password field with a generated password", () => {
    render(<NewUserDialog open={true} onClose={onClose} />);

    fireEvent.click(byId("new-user-generate-password"));

    expect(byId("new-user-password").value).toBe("Gen3rated!Pass");
    expect(submit()).toBeEnabled();
});

it("submits the user with extra attributes lifted into the extra object", async () => {
    vi.mocked(postNewUser).mockResolvedValue(undefined);

    render(<NewUserDialog open={true} onClose={onClose} />);

    fireEvent.change(byId("new-user-username"), { target: { value: "jane" } });
    fireEvent.change(byId("new-user-password"), { target: { value: "secret" } });
    fireEvent.click(screen.getByText("Show Additional Fields"));
    fireEvent.change(byId("new-user-display_name"), { target: { value: "Jane" } });
    fireEvent.change(byId("new-user-custom_field"), { target: { value: "x" } });

    await act(async () => {
        fireEvent.click(submit());
    });

    await waitFor(() => expect(postNewUser).toHaveBeenCalledOnce());

    const body = vi.mocked(postNewUser).mock.calls[0][0] as any;
    expect(body).toMatchObject({
        display_name: "Jane",
        extra: { custom_field: "x" },
        password: "secret",
        username: "jane",
    });
    expect(body).not.toHaveProperty("custom_field");

    expect(mockCreateSuccess).toHaveBeenCalledWith("User created successfully.");
    expect(onClose).toHaveBeenCalledOnce();
});

it("rejects an invalid username without calling the API", async () => {
    render(<NewUserDialog open={true} onClose={onClose} />);

    fireEvent.change(byId("new-user-username"), { target: { value: "bad user!" } });
    fireEvent.change(byId("new-user-password"), { target: { value: "secret" } });

    await act(async () => {
        fireEvent.click(submit());
    });

    expect(await screen.findByText(/Usernames must contain only alphanumeric characters/)).toBeInTheDocument();
    expect(postNewUser).not.toHaveBeenCalled();
});

it("notifies an error and keeps the dialog open when creation fails", async () => {
    vi.mocked(postNewUser).mockRejectedValue(new Error("boom"));

    render(<NewUserDialog open={true} onClose={onClose} />);

    fireEvent.change(byId("new-user-username"), { target: { value: "jane" } });
    fireEvent.change(byId("new-user-password"), { target: { value: "secret" } });

    await act(async () => {
        fireEvent.click(submit());
    });

    await waitFor(() => expect(mockCreateError).toHaveBeenCalledWith("Error creating user"));
    expect(onClose).not.toHaveBeenCalled();
    expect(byId("new-user-username").value).toBe("jane");
});

it("closes when cancel is clicked", () => {
    render(<NewUserDialog open={true} onClose={onClose} />);

    fireEvent.click(byId("new-user-cancel"));

    expect(onClose).toHaveBeenCalledOnce();
    expect(postNewUser).not.toHaveBeenCalled();
});

it("does not render the form while the metadata is loading", () => {
    mocks.metadata = undefined;
    mocks.metadataLoading = true;

    render(<NewUserDialog open={true} onClose={onClose} />);

    expect(byId("new-user-username")).toBeNull();
    expect(byId("new-user-submit")).toBeNull();
});

it("shows an error when the metadata fails to load", () => {
    mocks.metadata = undefined;
    mocks.metadataError = new Error("boom");

    render(<NewUserDialog open={true} onClose={onClose} />);

    expect(screen.getByText(/boom/)).toBeInTheDocument();
    expect(byId("new-user-submit")).toBeNull();
});
