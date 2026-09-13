import { DeleteWithOptionalResponse, Get, PostWithOptionalResponse } from "@services/Client";
import { deleteGroup, getAllGroups, postNewGroup } from "@services/GroupManagement";

vi.mock("@services/Api", () => ({
    AdminGroupRestPath: "/api/admin/groups",
}));

vi.mock("@services/Client", () => ({
    DeleteWithOptionalResponse: vi.fn(),
    Get: vi.fn(),
    PostWithOptionalResponse: vi.fn(),
}));

beforeEach(() => {
    vi.clearAllMocks();
});

it("gets all groups", async () => {
    vi.mocked(Get).mockResolvedValue(["admins", "dev"]);

    await expect(getAllGroups()).resolves.toEqual(["admins", "dev"]);
    expect(Get).toHaveBeenCalledWith("/api/admin/groups");
});

it("posts a new group", async () => {
    await postNewGroup({ name: "dev" });

    expect(PostWithOptionalResponse).toHaveBeenCalledWith("/api/admin/groups", { name: "dev" });
});

it("deletes a group", async () => {
    await deleteGroup("dev");

    expect(DeleteWithOptionalResponse).toHaveBeenCalledWith("/api/admin/groups/dev");
});
