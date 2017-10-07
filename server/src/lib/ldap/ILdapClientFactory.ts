
import { ILdapClient } from "./ILdapClient";

export interface ILdapClientFactory {
  create(): ILdapClient;
}