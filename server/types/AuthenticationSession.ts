import U2f = require("u2f");
import { WhitelistValue } from "../src/lib/authentication/whitelist/WhitelistHandler";

export interface AuthenticationSession {
  userid: string;
  first_factor: boolean;
  second_factor: boolean;
  whitelisted: WhitelistValue;
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