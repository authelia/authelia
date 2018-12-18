
export interface IdentityValidationDocument {
    userId: string;
    token: string;
    challenge: string;
    maxDate: Date;
}