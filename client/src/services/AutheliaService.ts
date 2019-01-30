import RemoteState from "../views/AuthenticationView/RemoteState";
import u2fApi, { SignRequest } from "u2f-api";

async function fetchSafe(url: string, options?: RequestInit) {
  return fetch(url, options)
    .then(async (res) => {
      if (res.status !== 200 && res.status !== 204) {
        throw new Error('Status code ' + res.status);
      }
      return res;
    });
}

/**
 * Fetch current authentication state.
 */
export async function fetchState() {
  return fetchSafe('/api/state')
    .then(async (res) => {
      const body = await res.json() as RemoteState;
      return body;
    });
}

export async function postFirstFactorAuth(username: string, password: string) {
  return fetchSafe('/api/firstfactor', {
    method: 'POST',
    headers: {
      'Accept': 'application/json',
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      username: username,
      password: password,
    })
  });
}

export async function postLogout() {
  return fetchSafe('/api/logout', {
    method: 'POST',
    headers: {
      'Accept': 'application/json',
      'Content-Type': 'application/json',
    },
  })
}

export async function startU2FRegistrationIdentityProcess() {
  return fetchSafe('/api/secondfactor/u2f/identity/start', {
    method: 'POST',
  });
}

export async function startTOTPRegistrationIdentityProcess() {
  return fetchSafe('/api/secondfactor/totp/identity/start', {
    method: 'POST',
  });
}

export async function requestSigning() {
  return fetchSafe('/api/u2f/sign_request')
    .then(async (res) => {
      const body = await res.json();
      return body as SignRequest;
    });
}

export async function completeSecurityKeySigning(response: u2fApi.SignResponse) {
  return fetchSafe('/api/u2f/sign', {
    method: 'POST',
    headers: {
      'Accept': 'application/json',
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(response),
  });
}

export async function verifyTotpToken(token: string) {
  return fetchSafe('/api/totp', {
    method: 'POST',
    headers: {
      'Accept': 'application/json',
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({token}),
  })
}

export async function initiatePasswordResetIdentityValidation(username: string) {
  return fetchSafe('/api/password-reset/identity/start', {
    method: 'POST',
    headers: {
      'Accept': 'application/json',
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({username})
  });
}

export async function completePasswordResetIdentityValidation(token: string) {
  return fetch(`/api/password-reset/identity/finish?token=${token}`, {
    method: 'POST',
  });
}

export async function resetPassword(newPassword: string) {
  return fetchSafe('/api/password-reset', {
    method: 'POST',
    headers: {
      'Accept': 'application/json',
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({password: newPassword})
  });
}

export async function checkRedirection(url: string) {
  const res = await fetch('/api/redirect', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({url})
  })

  if (res.status !== 200) {
    throw new Error('Status code ' + res.status);
  }

  const text = await res.text();
  if (text !== 'OK') {
    throw new Error('Cannot redirect');
  }
  return;
}