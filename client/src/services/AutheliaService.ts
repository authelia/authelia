import RemoteState from "../views/AuthenticationView/RemoteState";
import u2fApi, { SignRequest } from "u2f-api";
import Method2FA from "../types/Method2FA";
import RedirectResponse from "./RedirectResponse";
import PreferedMethodResponse from "./PreferedMethodResponse";

class AutheliaService {
  static async fetchSafe(url: string, options?: RequestInit): Promise<Response> {
    const res = await fetch(url, options);
    if (res.status !== 200 && res.status !== 204) {
      throw new Error('Status code ' + res.status);
    }
    return res;
  }

  static async fetchSafeJson<T>(url: string, options?: RequestInit): Promise<T> {
    const res = await fetch(url, options);
    if (res.status !== 200) {
      throw new Error('Status code ' + res.status);
    }
    return await res.json();
  }

  /**
   * Fetch current authentication state.
   */
  static async fetchState(): Promise<RemoteState> {
    return await this.fetchSafeJson('/api/state')
  }

  static async postFirstFactorAuth(username: string, password: string,
    rememberMe: boolean, redirectionUrl: string | null) {

    const headers: Record<string, string> = {
      'Accept': 'application/json',
      'Content-Type': 'application/json',
    }

    if (redirectionUrl) {
      headers['X-Target-Url'] = redirectionUrl;
    }

    return this.fetchSafe('/api/firstfactor', {
      method: 'POST',
      headers: headers,
      body: JSON.stringify({
        username: username,
        password: password,
        keepMeLoggedIn: rememberMe,
      })
    });
  }

  static async postLogout() {
    return this.fetchSafe('/api/logout', {
      method: 'POST',
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json',
      },
    })
  }

  static async startU2FRegistrationIdentityProcess() {
    return this.fetchSafe('/api/secondfactor/u2f/identity/start', {
      method: 'POST',
    });
  }

  static async startTOTPRegistrationIdentityProcess() {
    return this.fetchSafe('/api/secondfactor/totp/identity/start', {
      method: 'POST',
    });
  }

  static async requestSigning() {
    return this.fetchSafe('/api/u2f/sign_request')
      .then(async (res) => {
        const body = await res.json();
        return body as SignRequest;
      });
  }

  static async completeSecurityKeySigning(
    response: u2fApi.SignResponse, redirectionUrl: string | null) {

    const headers: Record<string, string> = {
      'Accept': 'application/json',
      'Content-Type': 'application/json',
    }
    if (redirectionUrl) {
      headers['X-Target-Url'] = redirectionUrl;
    }
    return this.fetchSafe('/api/u2f/sign', {
      method: 'POST',
      headers: headers,
      body: JSON.stringify(response),
    });
  }

  static async verifyTotpToken(
    token: string, redirectionUrl: string | null) {
    
      const headers: Record<string, string> = {
      'Accept': 'application/json',
      'Content-Type': 'application/json',
    }
    if (redirectionUrl) {
      headers['X-Target-Url'] = redirectionUrl;
    }
    return this.fetchSafe('/api/totp', {
      method: 'POST',
      headers: headers,
      body: JSON.stringify({token}),
    })
  }

  static async triggerDuoPush(redirectionUrl: string | null): Promise<RedirectResponse | undefined> {    
    const headers: Record<string, string> = {
      'Accept': 'application/json',
      'Content-Type': 'application/json',
    }
    if (redirectionUrl) {
      headers['X-Target-Url'] = redirectionUrl;
    }
    const res = await this.fetchSafe('/api/duo-push', {
      method: 'POST',
      headers: headers,
    });

    if (res.status === 204) {
      return;
    }

    const body = await res.json();
    if ('error' in body) {
      throw new Error(body['error']);
    }
    return body;
  }

  static async initiatePasswordResetIdentityValidation(username: string) {
    return this.fetchSafe('/api/password-reset/identity/start', {
      method: 'POST',
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({username})
    });
  }

  static async completePasswordResetIdentityValidation(token: string) {
    return fetch(`/api/password-reset/identity/finish?token=${token}`, {
      method: 'POST',
    });
  }

  static async resetPassword(newPassword: string) {
    return this.fetchSafe('/api/password-reset', {
      method: 'POST',
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({password: newPassword})
    });
  }

  static async fetchPrefered2faMethod(): Promise<Method2FA> {
    const doc = await this.fetchSafeJson<PreferedMethodResponse>('/api/secondfactor/preferences');
    if (!doc) {
      throw new Error("No response.");
    }

    if (doc.error) {
      throw new Error(doc.error);
    }

    if (!doc.method) {
      throw new Error("No method.");
    }

    return doc.method;
  }

  static async setPrefered2faMethod(method: Method2FA): Promise<void> {
    await this.fetchSafe('/api/secondfactor/preferences', {
      method: 'POST',
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({method})
    });
  }

  static async getAvailable2faMethods(): Promise<Method2FA[]> {
    return await this.fetchSafeJson('/api/secondfactor/available');
  }

  static async completeSecurityKeyRegistration(response: u2fApi.RegisterResponse): Promise<Response> {
    return await this.fetchSafe('/api/u2f/register', {
      method: 'POST',
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(response),
    });
  }

  static async requestSecurityKeyRegistration() {
    return this.fetchSafeJson<u2fApi.RegisterRequest>('/api/u2f/register_request')
  }
}

export default AutheliaService;