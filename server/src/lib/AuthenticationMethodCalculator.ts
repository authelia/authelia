import { AuthenticationMethod, AuthenticationMethodsConfiguration } from "./configuration/Configuration";

export class AuthenticationMethodCalculator {
  private configuration: AuthenticationMethodsConfiguration;

  constructor(config: AuthenticationMethodsConfiguration) {
    this.configuration = config;
  }

  compute(subDomain: string): AuthenticationMethod {
    if (this.configuration
      && this.configuration.per_subdomain_methods
      && subDomain in this.configuration.per_subdomain_methods) {
      return this.configuration.per_subdomain_methods[subDomain];
    }
    return this.configuration.default_method;
  }
}