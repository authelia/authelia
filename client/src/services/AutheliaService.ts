import RemoteState from "../views/AuthenticationView/RemoteState";
import U2fApi from "u2f-api";
import Method2FA from "../types/Method2FA";
import { string } from "prop-types";

interface DataResponse<T> {
  status: "OK";
  data: T;
}

interface ErrorResponse {
  status: "KO";
  message: string;
}

type ServiceResponse<T> = DataResponse<T> | ErrorResponse;

class AutheliaService {
  static async fetchSafeJson<T>(url: string, options?: RequestInit): Promise<T> {
    const res = await fetch(url, options);
    if (res.status !== 200) {
      throw new Error('Status code ' + res.status);
    }
    const response: ServiceResponse<T> = await res.json();
    if (response.status == "OK") {
      return response.data;
    } else {
      throw new Error(response.message)
    }
  }

  /**
   * Fetch current authentication state.
   */
  static async fetchState(): Promise<RemoteState> {
    return await this.fetchSafeJson<RemoteState>('/api/state')
  }

  static async postFirstFactorAuth(username: string, password: string,
    rememberMe: boolean, targetURL: string | null) {

    const headers: Record<string, string> = {
      'Accept': 'application/json',
      'Content-Type': 'application/json',
    }

    const requestBody: {
      username: string,
      password: string,
      keepMeLoggedIn: boolean,
      targetURL?: string
    } = {
      username: username,
      password: password,
      keepMeLoggedIn: rememberMe,
    }

    if (targetURL) {
      requestBody.targetURL = targetURL;
    }

    return this.fetchSafeJson<{redirect: string}|undefined>('/api/firstfactor', {
      method: 'POST',
      headers: headers,
      body: JSON.stringify(requestBody)
    });
  }

  static async postLogout() {
    return this.fetchSafeJson<undefined>('/api/logout', {
      method: 'POST',
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json',
      },
    })
  }

  static async startU2FRegistrationIdentityProcess() {
    return this.fetchSafeJson<undefined>('/api/secondfactor/u2f/identity/start', {
      method: 'POST',
    });
  }

  static async startTOTPRegistrationIdentityProcess() {
    return this.fetchSafeJson<undefined>('/api/secondfactor/totp/identity/start', {
      method: 'POST',
    });
  }

  static async requestSigning() {
    return this.fetchSafeJson<{
      appId: string,
      challenge: string,
      registeredKeys: {
        appId: string,
        keyHandle: string,
        version: string,
      }[]
    }>('/api/secondfactor/u2f/sign_request', {
      method: 'POST'
    });
  }

  static async completeSecurityKeySigning(
    response: U2fApi.SignResponse, targetURL: string | null) {

    const headers: Record<string, string> = {'Content-Type': 'application/json',}
    const requestBody: {signResponse: U2fApi.SignResponse, targetURL?: string} = {
      signResponse: response,
    };
    if (targetURL) {
      requestBody.targetURL = targetURL;
    }
    return this.fetchSafeJson<{redirect: string}|undefined>('/api/secondfactor/u2f/sign', {
      method: 'POST',
      headers: headers,
      body: JSON.stringify(requestBody),
    });
  }

  static async verifyTotpToken(
    token: string, targetURL: string | null) {    
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    }
    var requestBody: {token: string, targetURL?: string} = {token};
    if (targetURL) {
      requestBody.targetURL = targetURL;
    }
    return this.fetchSafeJson<{redirect: string}|undefined>('/api/secondfactor/totp', {
      method: 'POST',
      headers: headers,
      body: JSON.stringify(requestBody),
    })
  }

  static async triggerDuoPush(targetURL: string | null): Promise<{redirect: string}|undefined> {    
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    }
    const requestBody: {targetURL?: string} = {}
    if (targetURL) {
      requestBody.targetURL = targetURL;
    }
    return this.fetchSafeJson<{redirect: string}|undefined>('/api/secondfactor/duo', {
      method: 'POST',
      headers: headers,
      body: JSON.stringify(requestBody),
    });
  }

  static async initiatePasswordResetIdentityValidation(username: string) {
    return this.fetchSafeJson<undefined>('/api/reset-password/identity/start', {
      method: 'POST',
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({username})
    });
  }

  static async completePasswordResetIdentityValidation(token: string) {
    return this.fetchSafeJson<undefined>(`/api/reset-password/identity/finish`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({token})
    });
  }

  static async resetPassword(newPassword: string) {
    return this.fetchSafeJson<undefined>('/api/reset-password', {
      method: 'POST',
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({password: newPassword})
    });
  }

  static async fetchPrefered2faMethod(): Promise<Method2FA> {
    const res = await this.fetchSafeJson<{method: Method2FA}>('/api/secondfactor/preferences');
    return res.method;
  }

  static async setPrefered2faMethod(method: Method2FA): Promise<void> {
    return this.fetchSafeJson<undefined>('/api/secondfactor/preferences', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({method})
    });
  }

  static async getAvailable2faMethods(): Promise<Method2FA[]> {
    return this.fetchSafeJson('/api/secondfactor/available');
  }

  static async completeSecurityKeyRegistration(
    response: U2fApi.RegisterResponse): Promise<undefined> {
    return this.fetchSafeJson('/api/secondfactor/u2f/register', {
      method: 'POST',
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(response),
    });
  }

  static async completeSecurityKeyRegistrationIdentityValidation(token: string) {
    return this.fetchSafeJson<{
      appId: string,
      registerRequests: [{
        version: string,
        challenge: string,
      }]
    }>(`/api/secondfactor/u2f/identity/finish`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({token})
    });
  }

  static async completeOneTimePasswordRegistrationIdentityValidation(token: string) {
    return this.fetchSafeJson<{base32_secret: string, otpauth_url: string}>(`/api/secondfactor/totp/identity/finish`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({token})
    });
  }
}

export default AutheliaService;