

declare module "u2f" {
    export interface Request {
        version: "U2F_V2";
        appId: string;
        challenge: string;
        keyHandle?: string;
    }

    export interface RegistrationData {
        clientData: string;
        registrationData: string;
        errorCode?: number;
    }

    export interface RegistrationResult {
        successful: boolean;
        publicKey: string;
        keyHandle: string;
        certificate: string;
    }


    export interface SignatureData {
        clientData: string;
        signatureData: string;
        errorCode?: number;
    }

    export interface SignatureResult {
        successful: boolean;
        userPresent: boolean;
        counter: number;
    }

    export interface Error {
        errorCode: number;
        errorMessage: string;
    }

    export function request(appId: string, keyHandle?: string): Request;
    export function checkRegistration(request: Request, registerData: RegistrationData): RegistrationResult | Error;
    export function checkSignature(request: Request, signData: SignatureData, publicKey: string): SignatureResult | Error;
}