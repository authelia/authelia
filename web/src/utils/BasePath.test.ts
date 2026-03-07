import { getBasePath } from "@utils/BasePath";

beforeEach(() => {
    document.body.getAttributeNames().forEach((attr) => document.body.removeAttribute(attr));
});

it("returns the base path from the embedded variable", () => {
    document.body.setAttribute("data-basepath", "/auth");
    expect(getBasePath()).toBe("/auth");
});

it("throws when the base path is not set", () => {
    expect(() => getBasePath()).toThrow("No basepath embedded variable detected");
});
