import {
    ValidateDisplayName,
    ValidateEmail,
    ValidateGroup,
    ValidateUsername,
    getAttributeMetadata,
    isAttributeRequired,
    validateAttributeValue,
} from "@models/UserManagement";
import { UserAttributeMetadataBody } from "@services/UserManagement";

const metadata: UserAttributeMetadataBody = {
    required_attributes: ["username", "mail"],
    supported_attributes: {
        birthdate: { type: "date" },
        display_name: { type: "text" },
        groups: { multiple: true, type: "groups" },
        mail: { multiple: true, type: "email" },
        username: { type: "text" },
    },
};

it("validates usernames", () => {
    expect(ValidateUsername("john")).toBe(true);
    expect(ValidateUsername("john_doe-1,2")).toBe(true);
    expect(ValidateUsername("john@example.com")).toBe(true);
    expect(ValidateUsername("")).toBe(false);
    expect(ValidateUsername("   ")).toBe(false);
    expect(ValidateUsername("john doe")).toBe(false);
    expect(ValidateUsername("not@valid")).toBe(false);
    expect(ValidateUsername("a".repeat(101))).toBe(false);
});

it("validates display names", () => {
    expect(ValidateDisplayName("John Doe")).toBe(true);
    expect(ValidateDisplayName("Jöhn Døe 🙂")).toBe(true);
    expect(ValidateDisplayName("")).toBe(false);
    expect(ValidateDisplayName("  ")).toBe(false);
    expect(ValidateDisplayName("a".repeat(101))).toBe(false);
});

it("validates emails", () => {
    expect(ValidateEmail("john@example.com")).toBe(true);
    expect(ValidateEmail("john+tag@sub.example.co")).toBe(true);
    expect(ValidateEmail("")).toBe(false);
    expect(ValidateEmail("john")).toBe(false);
    expect(ValidateEmail("john@")).toBe(false);
    expect(ValidateEmail("john@example")).toBe(false);
});

it("validates groups", () => {
    expect(ValidateGroup("admins")).toBe(true);
    expect(ValidateGroup("dev-team_1")).toBe(true);
    expect(ValidateGroup("")).toBe(false);
    expect(ValidateGroup(" ")).toBe(false);
    expect(ValidateGroup("dev team")).toBe(false);
});

it("reports required attributes from metadata", () => {
    expect(isAttributeRequired("username", metadata)).toBe(true);
    expect(isAttributeRequired("display_name", metadata)).toBe(false);
    expect(isAttributeRequired("unknown", metadata)).toBe(false);
});

it("looks up attribute metadata", () => {
    expect(getAttributeMetadata("mail", metadata)).toEqual({ multiple: true, type: "email" });
    expect(getAttributeMetadata("unknown", metadata)).toBeUndefined();
});

it("accepts empty values for any attribute type", () => {
    expect(validateAttributeValue("", { type: "email" })).toBeNull();
    expect(validateAttributeValue(undefined, { type: "url" })).toBeNull();
    expect(validateAttributeValue(null, { type: "date" })).toBeNull();
});

it("validates single and multiple email attributes", () => {
    expect(validateAttributeValue("john@example.com", { type: "email" })).toBeNull();
    expect(validateAttributeValue("john", { type: "email" })).toBe("Invalid email format");
    expect(validateAttributeValue(["a@b.co", "c@d.co"], { multiple: true, type: "email" })).toBeNull();
    expect(validateAttributeValue(["a@b.co", "nope"], { multiple: true, type: "email" })).toBe("Invalid email format");
});

it("validates url attributes", () => {
    expect(validateAttributeValue("https://example.com", { type: "url" })).toBeNull();
    expect(validateAttributeValue("not a url", { type: "url" })).toBe("Invalid URL format");
});

it("validates telephone attributes", () => {
    expect(validateAttributeValue("+1 555 1234567", { type: "tel" })).toBeNull();
    expect(validateAttributeValue("(555) 123-4567", { type: "tel" })).toBeNull();
    expect(validateAttributeValue("abc", { type: "tel" })).toBe("Invalid phone number format");
});

it("validates date attributes", () => {
    expect(validateAttributeValue("2020-01-31", { type: "date" })).toBeNull();
    expect(validateAttributeValue("not a date", { type: "date" })).toBe("Invalid date format");
});

it("validates checkbox attributes", () => {
    expect(validateAttributeValue(true, { type: "checkbox" })).toBeNull();
    expect(validateAttributeValue("yes", { type: "checkbox" })).toBe("Invalid boolean value");
});

it("does not validate plain text attributes", () => {
    expect(validateAttributeValue("anything goes", { type: "text" })).toBeNull();
});
