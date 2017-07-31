
import * as ObjectPath from "object-path";
import {
  AppConfiguration, UserConfiguration, NotifierConfiguration,
  ACLConfiguration, LdapConfiguration, SessionRedisOptions,
  MongoStorageConfiguration, LocalStorageConfiguration
} from "./Configuration";

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

function adaptFromUserConfiguration(userConfiguration: UserConfiguration): AppConfiguration {
  ensure_key_existence(userConfiguration, "ldap");
  ensure_key_existence(userConfiguration, "session.secret");

  const port = ObjectPath.get(userConfiguration, "port", 8080);

  return {
    port: port,
    ldap: ObjectPath.get<object, LdapConfiguration>(userConfiguration, "ldap"),
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
    access_control: ObjectPath.get<object, ACLConfiguration>(userConfiguration, "access_control")
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

