import { FormatDateHumanReadable } from "@i18n/formats";

it("exports date format with expected properties", () => {
    expect(FormatDateHumanReadable).toEqual({
        day: "numeric",
        hour: "numeric",
        minute: "numeric",
        month: "long",
        year: "numeric",
    });
});
