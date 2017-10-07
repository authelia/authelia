
import * as ObjectPath from "object-path";
import {
  AppConfiguration, UserConfiguration, NotifierConfiguration,
  ACLConfiguration, LdapConfiguration, SessionRedisOptions,
  MongoStorageConfiguration, LocalStorageConfiguration,
  UserLdapConfiguration
} from "./Configuration";
import Util = require("util");

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
  const DEFAULT_GROUPS_FILTER =
    userConfig.additional_users_dn
      ? Util.format("member=cn={0},%s,%s", userConfig.additional_groups_dn, userConfig.base_dn)
      : Util.format("member=cn={0},%s", userConfig.base_dn);
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

function adaptFromUserConfiguration(userConfiguration: UserConfiguration): AppConfiguration {
  ensure_key_existence(userConfiguration, "ldap");
  // ensure_key_existence(userConfiguration, "ldap.url");
  // ensure_key_existence(userConfiguration, "ldap.base_dn");
  ensure_key_existence(userConfiguration, "session.secret");
  ensure_key_existence(userConfiguration, "regulation");

  const port = userConfiguration.port || 8080;
  const ldapConfiguration = adaptLdapConfiguration(userConfiguration.ldap);

  return {
    port: port,
    ldap: ldapConfiguration,
    session: {
      domain: ObjectPath.get<object, string>(userConfiguration, "session.domain"),
      secret: ObjectPath.get<object, string>(userConfiguration, "session.secret"),
      expiration: get_optional<number>(userConfiguration, "session.expiration", 3600000), // in ms
      redis: ObjectPath.get<object, SessionRedisOptions>(userConfiguration, "session.redis")
    },
    storage: {
      local: get_optional<LocalStorageConfiguration>(userConfiguration, "storage.local", undefined),
      mongo: get_optional<MongoStorageConfiguration>(userConfiguration, "storage.mongo", undefined)
    },
    logs_level: get_optional<string>(userConfiguration, "logs_level", "info"),
    notifier: ObjectPath.get<object, NotifierConfiguration>(userConfiguration, "notifier"),
    access_control: ObjectPath.get<object, ACLConfiguration>(userConfiguration, "access_control"),
    regulation: userConfiguration.regulation
  };
}

export class ConfigurationAdapter {
  static adapt(userConfiguration: UserConfiguration): AppConfiguration {
    const appConfiguration = adaptFromUserConfiguration(userConfiguration);

    const ldapUrl = process.env[LDAP_URL_ENV_VARIABLE];
    if (ldapUrl)
      appConfiguration.ldap.url = ldapUrl;

    return appConfiguration;
  }
}

