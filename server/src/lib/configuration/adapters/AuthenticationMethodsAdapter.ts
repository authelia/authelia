import { AuthenticationMethodsConfiguration } from "../Configuration";
import { ObjectCloner } from "../../utils/ObjectCloner";

function clone(obj: any): any {
  return JSON.parse(JSON.stringify(obj));
}

export class AuthenticationMethodsAdapter {
  static adapt(authentication_methods: AuthenticationMethodsConfiguration)
    : AuthenticationMethodsConfiguration {
    if (!authentication_methods) {
      return {
        default_method: "two_factor",
        per_subdomain_methods: {}
      };
    }

    const newAuthMethods: AuthenticationMethodsConfiguration
      = ObjectCloner.clone(authentication_methods);

    if (!newAuthMethods.default_method)
      newAuthMethods.default_method = "two_factor";

    if (!newAuthMethods.per_subdomain_methods ||
      newAuthMethods.per_subdomain_methods.constructor !== Object)
      newAuthMethods.per_subdomain_methods = {};

    return newAuthMethods;
  }
}
