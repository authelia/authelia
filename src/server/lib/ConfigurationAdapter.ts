
import * as ObjectPath from "object-path";
import { AppConfiguration, UserConfiguration, NotifierConfiguration, ACLConfiguration, LdapConfiguration } from "./../../types/Configuration";


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

export default class ConfigurationAdapter {
  static adapt(yaml_config: UserConfiguration): AppConfiguration {
    ensure_key_existence(yaml_config, "ldap");
    ensure_key_existence(yaml_config, "session.secret");

    const port = ObjectPath.get(yaml_config, "port", 8080);

    return {
      port: port,
      ldap: ObjectPath.get<object, LdapConfiguration>(yaml_config, "ldap"),
      session: {
        domain: ObjectPath.get<object, string>(yaml_config, "session.domain"),
        secret: ObjectPath.get<object, string>(yaml_config, "session.secret"),
        expiration: get_optional<number>(yaml_config, "session.expiration", 3600000), // in ms
      },
      store_directory: get_optional<string>(yaml_config, "store_directory", undefined),
      logs_level: get_optional<string>(yaml_config, "logs_level", "info"),
      notifier: ObjectPath.get<object, NotifierConfiguration>(yaml_config, "notifier"),
      access_control: ObjectPath.get<object, ACLConfiguration>(yaml_config, "access_control")
    };
  }
}

