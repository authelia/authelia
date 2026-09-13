import { DeleteWithOptionalResponse, Get, PatchWithOptionalResponse, PostWithOptionalResponse } from "@services/Client";
import {
    deleteUser,
    getAdminConfiguration,
    getAllUserInfo,
    getUser,
    getUserAttributeMetadata,
    patchChangeUser,
    postChangePasswordForUser,
    postNewUser,
    postSendResetPasswordEmailForUser,
} from "@services/UserManagement";

vi.mock("@services/Api", () => ({
    AdminChangePasswordRestSubPath: "/password/change",
    AdminConfigPath: "/api/admin/config",
    AdminSendResetPasswordEmailRestSubPath: "/password/reset",
    AdminUserAttributeMetadataPath: "/api/admin/user-fields",
    AdminUserRestPath: "/api/admin/users",
}));

vi.mock("@services/Client", () => ({
    DeleteWithOptionalResponse: vi.fn(),
    Get: vi.fn(),
    PatchWithOptionalResponse: vi.fn(),
    PostWithOptionalResponse: vi.fn(),
}));

beforeEach(() => {
    vi.clearAllMocks();
});

it("gets the admin configuration", async () => {
    const config = {
        admin_group: "admins",
        allow_admins_to_add_admins: false,
        enabled: true,
        group_management_enabled: true,
    };
    vi.mocked(Get).mockResolvedValue(config);

    await expect(getAdminConfiguration()).resolves.toEqual(config);
    expect(Get).toHaveBeenCalledWith("/api/admin/config");
});

it("gets the user attribute metadata", async () => {
    const body = { required_attributes: ["username"], supported_attributes: { username: { type: "text" } } };
    vi.mocked(Get).mockResolvedValue(body);

    await expect(getUserAttributeMetadata()).resolves.toEqual(body);
    expect(Get).toHaveBeenCalledWith("/api/admin/user-fields");
});

it("gets all users", async () => {
    const users = [{ username: "john" }];
    vi.mocked(Get).mockResolvedValue(users);

    await expect(getAllUserInfo()).resolves.toEqual(users);
    expect(Get).toHaveBeenCalledWith("/api/admin/users");
});

it("gets a single user", async () => {
    vi.mocked(Get).mockResolvedValue({ username: "john" });

    await expect(getUser("john")).resolves.toEqual({ username: "john" });
    expect(Get).toHaveBeenCalledWith("/api/admin/users/john");
});

it("patches a user with the update mask as a query parameter", async () => {
    await patchChangeUser("john", { display_name: "John", mail: ["john@example.com"] }, ["display_name", "mail"]);

    expect(PatchWithOptionalResponse).toHaveBeenCalledWith("/api/admin/users/john?update_mask=display_name,mail", {
        display_name: "John",
        mail: ["john@example.com"],
    });
});

it("drops an empty password from the patch body", async () => {
    await patchChangeUser("john", { display_name: "John", password: "" }, ["display_name"]);

    expect(PatchWithOptionalResponse).toHaveBeenCalledWith("/api/admin/users/john?update_mask=display_name", {
        display_name: "John",
    });
});

it("keeps a non-empty password in the patch body", async () => {
    await patchChangeUser("john", { password: "secret" }, ["password"]);

    expect(PatchWithOptionalResponse).toHaveBeenCalledWith("/api/admin/users/john?update_mask=password", {
        password: "secret",
    });
});

it("posts a new user", async () => {
    const user = { password: "secret", username: "jane" };

    await postNewUser(user);

    expect(PostWithOptionalResponse).toHaveBeenCalledWith("/api/admin/users", user);
});

it("deletes a user", async () => {
    await deleteUser("jane");

    expect(DeleteWithOptionalResponse).toHaveBeenCalledWith("/api/admin/users/jane");
});

it("changes a user password", async () => {
    await postChangePasswordForUser("jane", "secret");

    expect(PostWithOptionalResponse).toHaveBeenCalledWith("/api/admin/users/jane/password/change", {
        password: "secret",
    });
});

it("sends a password reset email", async () => {
    await postSendResetPasswordEmailForUser("jane");

    expect(PostWithOptionalResponse).toHaveBeenCalledWith("/api/admin/users/jane/password/reset");
});
