import U2f = require("u2f");

export interface AuthenticationSession {
  userid: string;
  first_factor: boolean;
  second_factor: boolean;
  whitelisted: boolean;
  last_activity_datetime: number;
  identity_check?: {
    challenge: string;
    userid: string;
  };
  register_request?: U2f.Request;
  sign_request?: U2f.Request;
  email: string;
  groups: string[];
  redirect?: string;
}