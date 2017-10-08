import { ACLConfiguration } from "../Configuration";

function clone(obj: any): any {
  return JSON.parse(JSON.stringify(obj));
}

const DEFAULT_POLICY = "deny";

function adaptDefaultPolicy(configuration: ACLConfiguration) {
  if (!configuration.default_policy)
    configuration.default_policy = DEFAULT_POLICY;
  if (configuration.default_policy != "deny" && configuration.default_policy != "allow")
    configuration.default_policy = DEFAULT_POLICY;
}

function adaptAny(configuration: ACLConfiguration) {
  if (!configuration.any || !(configuration.any.constructor === Array))
    configuration.any = [];
}

function adaptGroups(configuration: ACLConfiguration) {
  if (!configuration.groups || !(configuration.groups.constructor === Object))
    configuration.groups = {};
}

function adaptUsers(configuration: ACLConfiguration) {
  if (!configuration.users || !(configuration.users.constructor === Object))
    configuration.users = {};
}

export class ACLAdapter {
  static adapt(configuration: ACLConfiguration): ACLConfiguration {
    if (!configuration) return;

    const newConfiguration: ACLConfiguration = clone(configuration);
    adaptDefaultPolicy(newConfiguration);
    adaptAny(newConfiguration);
    adaptGroups(newConfiguration);
    adaptUsers(newConfiguration);
    return newConfiguration;
  }
}