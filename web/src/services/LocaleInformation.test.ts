import { Get } from "@services/Client";
import { getLocaleInformation } from "@services/LocaleInformation";

beforeEach(() => {
    vi.spyOn(console, "error").mockImplementation(() => {});
});

vi.mock("@services/Api", () => ({
    LocaleInformationPath: "/locales",
}));
vi.mock("@services/Client", () => ({
    Get: vi.fn(),
}));

it("gets locale information successfully", async () => {
    const mockData = { defaults: { language: "en", namespace: "common" }, languages: ["en"], namespaces: ["common"] };
    (Get as any).mockResolvedValue(mockData);
    const result = await getLocaleInformation();
    expect(Get).toHaveBeenCalledWith("/locales");
    expect(result).toEqual(mockData);
});

it("throws on get failure", async () => {
    (Get as any).mockRejectedValue(new Error("network error"));
    await expect(getLocaleInformation()).rejects.toThrow("Failed to fetch locale information");
});
