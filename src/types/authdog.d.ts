
declare module "authdog" {
    interface RegisterRequest {
        challenge: string;
    }

    interface RegisteredKey {
        version: number;
        keyHandle: string;
    }

    type RegisteredKeys = Array<RegisteredKey>;
    type RegisterRequests = Array<RegisterRequest>;
    type AppId = string;

    interface RegistrationRequest {
        appId: AppId;
        type: string;
        registerRequests: RegisterRequests;
        registeredKeys: RegisteredKeys;
    }

    interface Registration {
        publicKey: string;
        keyHandle: string;
        certificate: string;
    }

    interface ClientData {
        challenge: string;
    }

    interface RegistrationResponse {
        clientData: ClientData;
        registrationData: string;
    }

    interface Options {
        timeoutSeconds: number;
        requestId: string;
    }

    interface AuthenticationRequest {
        appId: AppId;
        type: string;
        challenge: string;
        registeredKeys: RegisteredKeys;
        timeoutSeconds: number;
        requestId: string;
    }

    interface AuthenticationResponse {
        keyHandle: string;
        clientData: ClientData;
        signatureData: string;
    }

    interface Authentication {
        userPresence: Uint8Array,
        counter: Uint32Array
    }

    export function startRegistration(appId: AppId, registeredKeys: RegisteredKeys, options?: Options): Promise<RegistrationRequest>;
    export function finishRegistration(registrationRequest: RegistrationRequest, registrationResponse: RegistrationResponse): Promise<Registration>;
    export function startAuthentication(appId: AppId, registeredKeys: RegisteredKeys, options: Options): Promise<AuthenticationRequest>;
    export function finishAuthentication(challenge: string, deviceResponse: AuthenticationResponse, registeredKeys: RegisteredKeys): Promise<Authentication>;
}