import {
  AuthenticationMethod,
  AuthenticationMethodsConfiguration
} from "../configuration/Configuration";

function computeIsSingleFactorOnlyMode(
  configuration: AuthenticationMethodsConfiguration): boolean {
  if (!configuration)
    return false;

  const method: AuthenticationMethod = configuration.default_method;
  if (configuration.default_method == "two_factor")
    return false;

  if (configuration.per_subdomain_methods) {
    for (const key in configuration.per_subdomain_methods) {
      const method = configuration.per_subdomain_methods[key];
      if (method == "two_factor")
        return false;
    }
  }
  return true;
}

export class MethodCalculator {
  static compute(config: AuthenticationMethodsConfiguration, subDomain: string)
    : AuthenticationMethod {
    if (config
      && config.per_subdomain_methods
      && subDomain in config.per_subdomain_methods) {
      return config.per_subdomain_methods[subDomain];
    }
    return config.default_method;
  }

  static isSingleFactorOnlyMode(config: AuthenticationMethodsConfiguration)
    : boolean {
    return computeIsSingleFactorOnlyMode(config);
  }
}