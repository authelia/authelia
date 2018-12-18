import { LdapConfiguration } from "./LdapConfiguration";
import { FileUsersDatabaseConfiguration } from "./FileUsersDatabaseConfiguration";

export interface AuthenticationBackendConfiguration {
  ldap?: LdapConfiguration;
  file?: FileUsersDatabaseConfiguration;
}

export function complete(
  configuration: AuthenticationBackendConfiguration)
  : [AuthenticationBackendConfiguration, string] {

  const newConfiguration: AuthenticationBackendConfiguration = (configuration)
    ? JSON.parse(JSON.stringify(configuration)) : {};

  if (Object.keys(newConfiguration).length != 1) {
    return [
      newConfiguration,
      "Authentication backend must have one of the following keys:" +
      "`ldap` or `file`"
    ];
  }

  return [newConfiguration, undefined];
}