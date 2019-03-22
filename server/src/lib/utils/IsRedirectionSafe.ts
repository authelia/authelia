import { ServerVariables } from "../ServerVariables";
import * as URLParse from "url-parse";

export default function IsRedirectionSafe(
  vars: ServerVariables,
  url: URLParse): boolean {

  const urlInDomain = url.hostname.endsWith(vars.config.session.domain);
  const protocolIsHttps = url.protocol === "https:";
  return urlInDomain && protocolIsHttps;
}
