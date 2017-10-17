
import * as ObjectPath from "object-path";
import {
  AppConfiguration, UserConfiguration, NotifierConfiguration,
  ACLConfiguration, LdapConfiguration, SessionRedisOptions,
  MongoStorageConfiguration, LocalStorageConfiguration,
  UserLdapConfiguration
} from "./Configuration";
import Util = require("util");
import { ACLAdapter } from "./adapters/ACLAdapter";
import { AuthenticationMethodsAdapter } from "./adapters/AuthenticationMethodsAdapter";
import { Validator } from "./Validator";

const LDAP_URL_ENV_VARIABLE = "LDAP_URL";

function get_optional<T>(config: object, path: string, default_value: T): T {
  let entry = default_value;
  if (ObjectPath.has(config, path)) {
    entry = ObjectPath.get<object, T>(config, path);
  }
  return entry;
}

function ensure_key_existence(config: object, path: string): void {
  if (!ObjectPath.has(config, path)) {
    throw new Error(`Configuration error: key '${path}' is missing in configuration file`);
  }
}

function adaptLdapConfiguration(userConfig: UserLdapConfiguration): LdapConfiguration {
  const DEFAULT_USERS_FILTER = "cn={0}";
  const DEFAULT_GROUPS_FILTER = "member={dn}";
  const DEFAULT_GROUP_NAME_ATTRIBUTE = "cn";
  const DEFAULT_MAIL_ATTRIBUTE = "mail";

  let usersDN = userConfig.base_dn;
  if (userConfig.additional_users_dn)
    usersDN = userConfig.additional_users_dn + "," + usersDN;

  let groupsDN = userConfig.base_dn;
  if (userConfig.additional_groups_dn)
    groupsDN = userConfig.additional_groups_dn + "," + groupsDN;

  return {
    url: userConfig.url,
    users_dn: usersDN,
    users_filter: userConfig.users_filter || DEFAULT_USERS_FILTER,
    groups_dn: groupsDN,
    groups_filter: userConfig.groups_filter || DEFAULT_GROUPS_FILTER,
    group_name_attribute: userConfig.group_name_attribute || DEFAULT_GROUP_NAME_ATTRIBUTE,
    mail_attribute: userConfig.mail_attribute || DEFAULT_MAIL_ATTRIBUTE,
    password: userConfig.password,
    user: userConfig.user
  };
}

function adaptFromUserConfiguration(userConfiguration: UserConfiguration)
  : AppConfiguration {
  if (!Validator.isValid(userConfiguration))
    throw new Error("Configuration is malformed. Please double check your configuration file.");

  const port = userConfiguration.port || 8080;
  const ldapConfiguration = adaptLdapConfiguration(userConfiguration.ldap);
  const authenticationMethods = AuthenticationMethodsAdapter
    .adapt(userConfiguration.authentication_methods);

  return {
    port: port,
    ldap: ldapConfiguration,
    session: {
      domain: ObjectPath.get<object, string>(userConfiguration, "session.domain"),
      secret: ObjectPath.get<object, string>(userConfiguration, "session.secret"),
      expiration: get_optional<number>(userConfiguration, "session.expiration", 3600000), // in ms
      inactivity: get_optional<number>(userConfiguration, "session.inactivity", undefined),
      redis: ObjectPath.get<object, SessionRedisOptions>(userConfiguration, "session.redis")
    },
    storage: {
      local: get_optional<LocalStorageConfiguration>(userConfiguration, "storage.local", undefined),
      mongo: get_optional<MongoStorageConfiguration>(userConfiguration, "storage.mongo", undefined)
    },
    logs_level: get_optional<string>(userConfiguration, "logs_level", "info"),
    notifier: ObjectPath.get<object, NotifierConfiguration>(userConfiguration, "notifier"),
    access_control: ACLAdapter.adapt(userConfiguration.access_control),
    regulation: userConfiguration.regulation,
    authentication_methods: authenticationMethods,
    default_redirection_url: userConfiguration.default_redirection_url
  };
}

export class ConfigurationParser {
  static parse(userConfiguration: UserConfiguration): AppConfiguration {
    const errors = Validator.isValid(userConfiguration);
    if (errors.length > 0) {
      errors.forEach((e: string) => { console.log(e); });
      throw new Error("Malformed configuration. Please double-check your configuration file.");
    }
    const appConfiguration = adaptFromUserConfiguration(userConfiguration);

    const ldapUrl = process.env[LDAP_URL_ENV_VARIABLE];
    if (ldapUrl)
      appConfiguration.ldap.url = ldapUrl;

    return appConfiguration;
  }
}

