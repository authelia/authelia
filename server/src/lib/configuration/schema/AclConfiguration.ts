
export type ACLPolicy = "deny" | "allow";

export type ACLRule = {
  domain: string;
  policy: ACLPolicy;
  whitelist_policy?: ACLPolicy;
  resources?: string[];
};

export type ACLDefaultRules = ACLRule[];
export type ACLGroupsRules = { [group: string]: ACLRule[]; };
export type ACLUsersRules = { [user: string]: ACLRule[]; };

export interface ACLConfiguration {
  default_policy?: ACLPolicy;
  default_whitelist_policy?: ACLPolicy;
  any?: ACLDefaultRules;
  groups?: ACLGroupsRules;
  users?: ACLUsersRules;
}

export function complete(configuration: ACLConfiguration): ACLConfiguration {
  const newConfiguration: ACLConfiguration = (configuration)
    ? JSON.parse(JSON.stringify(configuration)) : {};

  if (!newConfiguration.default_policy) {
    newConfiguration.default_policy = "allow";
  }

  if (!newConfiguration.default_whitelist_policy) {
    newConfiguration.default_whitelist_policy = "allow";
  }

  if (!newConfiguration.any) {
    newConfiguration.any = [];
  }

  if (!newConfiguration.groups) {
    newConfiguration.groups = {};
  }

  if (!newConfiguration.users) {
    newConfiguration.users = {};
  }

  return newConfiguration;
}