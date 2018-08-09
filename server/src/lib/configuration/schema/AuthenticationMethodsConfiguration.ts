export type AuthenticationMethod = "two_factor" | "single_factor";
export type AuthenticationMethodPerSubdomain = { [subdomain: string]: AuthenticationMethod };

export interface AuthenticationMethodsConfiguration {
  default_method?: AuthenticationMethod;
  per_subdomain_methods?: AuthenticationMethodPerSubdomain;
}

export function complete(configuration: AuthenticationMethodsConfiguration): AuthenticationMethodsConfiguration {
  const newConfiguration: AuthenticationMethodsConfiguration = (configuration) ? JSON.parse(JSON.stringify(configuration)) : {};

  if (!newConfiguration.default_method) {
    newConfiguration.default_method = "two_factor";
  }

  if (!newConfiguration.per_subdomain_methods) {
    newConfiguration.per_subdomain_methods = {};
  }

  return newConfiguration;
}