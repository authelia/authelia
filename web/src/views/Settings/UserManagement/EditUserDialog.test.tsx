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

// The backend already prefixes extra (non-standard) attribute names with "extra." in
// `supported_attributes` (see FileUserManagement.GetSupportedAttributes), so the fixture key
// below mirrors that shape - "custom_field" alone is not how a real backend ever names it.
const metadata = {
    required_attributes: ["username"],
    supported_attributes: {
        display_name: { type: "text" },
        "extra.custom_field": { type: "text" },
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

// A separate metadata/user fixture covering every attribute type UserFormField dispatches on.
// Kept isolated from `metadata`/`user` above (used only by its own test) because the "prefixes
// extra fields" test asserts the *entire* extra sub-object verbatim; adding more extra-typed
// attributes to the shared metadata would leak into that object and break the strict equality.
const allTypesMetadata = {
    required_attributes: ["username"],
    supported_attributes: {
        birthdate: { type: "date" },
        display_name: { type: "text" },
        "extra.backup_email": { type: "email" },
        "extra.custom_field": { type: "text" },
        // extra attribute with a name distinct from "birthdate" so it exercises the generic
        // type-based date dispatch rather than the name-based special-casing "birthdate" gets
        "extra.hire_date": { type: "date" },
        "extra.login_count": { type: "number" },
        "extra.newsletter": { type: "checkbox" },
        groups: { multiple: true, type: "groups" },
        locality: { type: "text" },
        mail: { multiple: true, type: "email" },
        password: { type: "password" },
        phone_number: { type: "tel" },
        username: { type: "text" },
        website: { type: "url" },
    },
};

const allTypesUser: UserDetailsExtended = {
    address: { locality: "Paris" },
    birthdate: "1990-01-01",
    display_name: "John",
    extra: {
        backup_email: "old@example.com",
        custom_field: "a",
        hire_date: "2020-01-01",
        login_count: 3,
        newsletter: false,
    },
    groups: ["dev"],
    mail: ["john@example.com"],
    phone_number: "+15551234567",
    username: "john",
    website: "https://example.com",
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

it("submits the correct JS type for every additional field type when edited", async () => {
    vi.mocked(patchChangeUser).mockResolvedValue(undefined);
    mocks.metadata = allTypesMetadata;

    render(<EditUserDialog open={true} user={allTypesUser} onClose={onClose} />);

    fireEvent.click(screen.getByText("Show Additional Fields"));

    fireEvent.change(byId("edit-user-phone_number"), { target: { value: "+442071234567" } });
    fireEvent.change(byId("edit-user-website"), { target: { value: "https://example.org" } });
    fireEvent.change(byId("edit-user-birthdate"), { target: { value: "1991-02-02" } });
    fireEvent.change(byId("edit-user-extra-hire_date"), { target: { value: "2024-05-01" } });
    fireEvent.click(byId("edit-user-extra-newsletter"));
    fireEvent.change(byId("edit-user-extra-login_count"), { target: { value: "9" } });
    fireEvent.change(byId("edit-user-extra-backup_email"), { target: { value: "new@example.com" } });

    await act(async () => {
        fireEvent.click(submit());
    });

    await waitFor(() => expect(patchChangeUser).toHaveBeenCalledOnce());

    const [username, body, mask] = vi.mocked(patchChangeUser).mock.calls[0];

    expect(username).toBe("john");
    expect(mask).toEqual(
        expect.arrayContaining([
            "phone_number",
            "website",
            "birthdate",
            "extra.hire_date",
            "extra.newsletter",
            "extra.login_count",
            "extra.backup_email",
        ]),
    );
    expect(mask).not.toContain("display_name");
    expect(mask).not.toContain("address.locality");
    expect(mask).not.toContain("groups");

    expect(body.phone_number).toBe("+442071234567");
    expect(body.website).toBe("https://example.org");
    expect(body.birthdate).toBe("1991-02-02");
    expect(body.extra).toEqual({
        backup_email: "new@example.com",
        custom_field: "a",
        hire_date: "2024-05-01",
        login_count: 9,
        newsletter: true,
    });
    expect(typeof body.extra!.login_count).toBe("number");
    expect(typeof body.extra!.newsletter).toBe("boolean");
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
