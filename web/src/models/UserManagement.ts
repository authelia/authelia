import { REGEX } from "@constants/Regex.js";
import { FieldMetadata, UserFieldMetadataBody } from "@services/UserManagement.js";

export interface UserDetailsExtended {
    username: string;
    password?: string; // Only used for new users.

    display_name?: string; // Optional, takes precedence over full_name for display
    emails?: string[]; //All Emails
    groups?: string[];

    first_name?: string; //GivenName
    last_name?: string; //FamilyName
    full_name?: string; //CommonName
    middle_name?: string; //MiddleName
    nickname?: string;
    profile?: string; //URL
    picture?: string; //URL
    website?: string; //URL
    gender?: string;
    birthdate?: string;
    zone_info?: string;
    locale?: string;
    phone_number?: string;
    phone_extension?: string;
    address?: UserDetailsAddress;

    extra?: Record<string, any>;

    last_logged_in?: string;
    last_password_change?: string;
    user_created_at?: string;
    method?: string;
    has_totp?: boolean;
    has_webauthn?: boolean;
    has_duo?: boolean;
}

export interface UserDetailsAddress {
    street_address?: string;
    locality?: string;
    region?: string;
    postal_code?: string;
    country?: string;
}

export interface CreateUserRequest extends Omit<
    UserDetailsExtended,
    "has_duo" | "has_totp" | "has_webauthn" | "last_logged_in" | "last_password_change" | "method" | "user_created_at"
> {
    username: string;
    password: string;
}

export function validateFieldValue(value: any, metadata: FieldMetadata, _fieldName: string): null | string {
    if (metadata.type === "email" && typeof value === "string") {
        if (!ValidateEmail(value)) {
            return `Invalid ${metadata.display_name.toLowerCase()} format`;
        }
    }

    if (metadata.type === "string" && typeof value === "string") {
        if (metadata.maxLength && value.length > metadata.maxLength) {
            return `${metadata.display_name} must be ${metadata.maxLength} characters or less`;
        }

        if (metadata.pattern) {
            const regex = new RegExp(metadata.pattern);
            if (!regex.test(value)) {
                return `Invalid ${metadata.display_name.toLowerCase()} format`;
            }
        }
    }

    return null;
}

export function isFieldRequired(fieldName: keyof CreateUserRequest, metadata: UserFieldMetadataBody): boolean {
    return metadata.required_fields.includes(fieldName);
}

export function getFieldMetadata(
    fieldName: keyof UserDetailsExtended,
    metadata: UserFieldMetadataBody,
): FieldMetadata | undefined {
    return metadata.field_metadata[fieldName];
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
