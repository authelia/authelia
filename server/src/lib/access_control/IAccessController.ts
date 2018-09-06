import { WhitelistValue } from "../authentication/whitelist/WhitelistHandler";

export interface IAccessController {
  isAccessAllowed(domain: string, resource: string, user: string, groups: string[], whitelisted: WhitelistValue, secondFactorAuth: boolean): boolean;
}