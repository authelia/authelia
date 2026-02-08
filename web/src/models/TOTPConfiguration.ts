export interface UserInfoTOTPConfiguration {
    created_at: Date;
    last_used_at?: Date;
    issuer: string;
    algorithm: TOTPAlgorithm;
    digits: TOTPDigits;
    period: number;
}

export interface TOTPOptions {
    algorithm: TOTPAlgorithm;
    algorithms: TOTPAlgorithm[];
    length: TOTPDigits;
    lengths: TOTPDigits[];
    period: number;
    periods: number[];
}

/* eslint-disable no-unused-vars */
export enum TOTPAlgorithm {
    SHA1 = 0,
    SHA256,
    SHA512,
}

export type TOTPDigits = 6 | 8;
export type TOTPAlgorithmPayload = "SHA1" | "SHA256" | "SHA512";

export function toAlgorithmString(alg: TOTPAlgorithm): TOTPAlgorithmPayload {
    switch (alg) {
        case TOTPAlgorithm.SHA1:
            return "SHA1";
        case TOTPAlgorithm.SHA256:
            return "SHA256";
        case TOTPAlgorithm.SHA512:
            return "SHA512";
    }
}

export function toEnum(alg: TOTPAlgorithmPayload): TOTPAlgorithm {
    switch (alg) {
        case "SHA1":
            return TOTPAlgorithm.SHA1;
        case "SHA256":
            return TOTPAlgorithm.SHA256;
        case "SHA512":
            return TOTPAlgorithm.SHA512;
    }
}
