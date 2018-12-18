import U2f = require("u2f");
import { Level } from "../src/lib/authentication/Level";

export interface AuthenticationSession {
  userid: string;
  authentication_level: Level;
  keep_me_logged_in: boolean;
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