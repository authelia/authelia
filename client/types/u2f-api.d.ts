

declare module "u2f-api" {
    type MessageTypes = "u2f_register_request" | "u2f_sign_request" | "u2f_register_response" | "u2f_sign_response";

    export interface Request {
        type: MessageTypes,
        signRequests: SignRequest[],
        registerRequests?: RegisterRequest[],
        timeoutSeconds?: number,
        requestId?: number
    }

    type ResponseData = Error | RegisterResponse | SignResponse;


    export interface Response {
        type: MessageTypes;
        responseData: ResponseData;
        requestId?: number;
    }

    export enum ErrorCodes {
        'OK' = 0,
        'OTHER_ERROR' = 1,
        'BAD_REQUEST' = 2,
        'CONFIGURATION_UNSUPPORTED' = 3,
        'DEVICE_INELIGIBLE' = 4,
        'TIMEOUT' = 5
    }

    export interface Error {
        errorCode: ErrorCodes;
        errorMessage?: string;
    }

    export interface RegisterResponse {
        registrationData: string;
        clientData: string;
    }

    export interface RegisterRequest {
        version: string;
        challenge: string;
        appId: string;
    }

    export interface SignResponse {
        keyHandle: string;
        signatureData: string;
        clientData: string;
    }

    export interface SignRequest {
        version: string;
        challenge: string;
        keyHandle: string;
        appId: string;
    }

    export function sign(signRequests: SignRequest[], timeout: number): Promise<SignResponse>;
    export function register(registerRequests: RegisterRequest[], signRequests: SignRequest[], timeout: number): Promise<RegisterResponse>;
}