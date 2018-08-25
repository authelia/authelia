import { ACLConfiguration, complete as AclConfigurationComplete } from "./AclConfiguration";
import { AuthenticationMethodsConfiguration, complete as AuthenticationMethodsConfigurationComplete } from "./AuthenticationMethodsConfiguration";
import { AuthenticationBackendConfiguration, complete as AuthenticationBackendComplete } from "./AuthenticationBackendConfiguration";
import { NotifierConfiguration, complete as NotifierConfigurationComplete } from "./NotifierConfiguration";
import { RegulationConfiguration, complete as RegulationConfigurationComplete } from "./RegulationConfiguration";
import { SessionConfiguration, complete as SessionConfigurationComplete } from "./SessionConfiguration";
import { StorageConfiguration, complete as StorageConfigurationComplete } from "./StorageConfiguration";
import { TotpConfiguration, complete as TotpConfigurationComplete } from "./TotpConfiguration";
import { MethodCalculator } from "../../authentication/MethodCalculator";

export interface Configuration {
  access_control?: ACLConfiguration;
  authentication_backend: AuthenticationBackendConfiguration;
  authentication_methods?: AuthenticationMethodsConfiguration;
  default_redirection_url?: string;
  logs_level?: string;
  notifier?: NotifierConfiguration;
  port?: number;
  regulation?: RegulationConfiguration;
  session?: SessionConfiguration;
  storage?: StorageConfiguration;
  totp?: TotpConfiguration;
}

export function complete(
  configuration: Configuration):
  [Configuration, string[]] {

  const newConfiguration: Configuration = JSON.parse(
    JSON.stringify(configuration));
  const errors: string[] = [];

  newConfiguration.access_control =
    AclConfigurationComplete(
      newConfiguration.access_control);

  const [backend, error] =
    AuthenticationBackendComplete(
      newConfiguration.authentication_backend);

  if (error) errors.push(error);
  newConfiguration.authentication_backend = backend;

  newConfiguration.authentication_methods =
    AuthenticationMethodsConfigurationComplete(
      newConfiguration.authentication_methods);

  if (!newConfiguration.logs_level) {
    newConfiguration.logs_level = "info";
  }

  // In single factor mode, notifier section is optional.
  if (!MethodCalculator.isSingleFactorOnlyMode(
      newConfiguration.authentication_methods) ||
      newConfiguration.notifier) {

    const [notifier, error] = NotifierConfigurationComplete(
      newConfiguration.notifier);
    newConfiguration.notifier = notifier;

    if (error) errors.push(error);
  }

  if (!newConfiguration.port) {
    newConfiguration.port = 8080;
  }

  newConfiguration.regulation = RegulationConfigurationComplete(
    newConfiguration.regulation);
  newConfiguration.session = SessionConfigurationComplete(
    newConfiguration.session);
  newConfiguration.storage = StorageConfigurationComplete(
    newConfiguration.storage);
  newConfiguration.totp = TotpConfigurationComplete(
    newConfiguration.totp);

  return [newConfiguration, errors];
}