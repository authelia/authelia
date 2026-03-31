import { REGEX } from "@constants/Regex.js";
import { AttributeMetadata, UserAttributeMetadataBody } from "@services/UserManagement.js";

export interface UserDetailsExtended {
    username: string;
    password?: string;
    groups?: string[];

    display_name?: string;
    mail?: string[];

    given_name?: string;
    family_name?: string;
    middle_name?: string;
    nickname?: string;
    profile?: string;
    picture?: string;
    website?: string;
    gender?: string;
    birthdate?: string;
    zoneinfo?: string;
    locale?: string;
    phone_number?: string;
    phone_extension?: string;
    street_address?: string;
    locality?: string;
    region?: string;
    postal_code?: string;
    country?: string;

    member_of?: string[];
    group_name?: string;
    group_member?: string[];

    extra?: Record<string, any>;

    // Read-only fields
    last_logged_in?: string;
    last_password_change?: string;
    user_created_at?: string;
    method?: string;
    has_totp?: boolean;
    has_webauthn?: boolean;
    has_duo?: boolean;
}

export interface CreateUserRequest extends Omit<
    UserDetailsExtended,
    "has_duo" | "has_totp" | "has_webauthn" | "last_logged_in" | "last_password_change" | "method" | "user_created_at"
> {
    username: string;
    password: string;
}

export function validateAttributeValue(value: any, metadata: AttributeMetadata): null | string {
    if (!value) return null;

    if (metadata.type === "email") {
        if (metadata.multiple && Array.isArray(value)) {
            for (const email of value) {
                if (typeof email === "string" && !ValidateEmail(email)) {
                    return `Invalid email format`;
                }
            }
        } else if (typeof value === "string" && !ValidateEmail(value)) {
            return `Invalid email format`;
        }
    }

    if (metadata.type === "url" && typeof value === "string") {
        try {
            new URL(value);
        } catch {
            return `Invalid URL format`;
        }
    }

    if (metadata.type === "tel" && typeof value === "string") {
        //TODO: validate this regex is correct for international phone numbers, including those with country codes.
        if (!/^[+]?[(]?[0-9]{1,4}[)]?[-\s.]?[(]?[0-9]{1,4}[)]?[-\s.]?[0-9]{1,9}$/im.test(value)) {
            return `Invalid phone number format`;
        }
    }

    if (metadata.type === "date" && typeof value === "string") {
        if (isNaN(Date.parse(value))) {
            return `Invalid date format`;
        }
    }

    if (metadata.type === "checkbox" && typeof value !== "boolean") {
        return `Invalid boolean value`;
    }

    return null;
}

export function isAttributeRequired(fieldName: string, metadata: UserAttributeMetadataBody): boolean {
    return metadata.required_attributes.includes(fieldName);
}

export function getAttributeMetadata(
    fieldName: string,
    metadata: UserAttributeMetadataBody,
): AttributeMetadata | undefined {
    return metadata.supported_attributes[fieldName];
}

export function ValidateUsername(username: string) {
    if (!username) return false;

    if (!username.trim()) return false;

    if (username.includes("@")) {
        return REGEX.EMAIL.test(username);
    }
    return REGEX.USERNAME.test(username);
}

export function ValidateDisplayName(displayName: string) {
    if (!displayName) return false;

    if (!displayName.trim()) return false;

    return REGEX.DISPLAY_NAME.test(displayName);
}

export function ValidateEmail(email: string) {
    if (!email.trim()) return false;

    return REGEX.EMAIL.test(email);
}

export function ValidateGroup(group: string) {
    if (!group) return false;

    if (!group.trim()) return false;

    return REGEX.GROUP.test(group);
}
