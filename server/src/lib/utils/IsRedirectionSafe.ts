import { ServerVariables } from "../ServerVariables";
import { Level } from "../authentication/Level";
import * as URLParse from "url-parse";
import { AuthenticationSession } from "AuthenticationSession";

export default function IsRedirectionSafe(
  vars: ServerVariables,
  authSession: AuthenticationSession,
  url: URLParse): boolean {

  const urlInDomain = url.hostname.endsWith(vars.config.session.domain);
  const protocolIsHttps = url.protocol === "https:";
  return urlInDomain && protocolIsHttps;
}
