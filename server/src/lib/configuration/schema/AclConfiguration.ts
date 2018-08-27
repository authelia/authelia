
export type ACLPolicy = "deny" | "allow";

export type ACLRule = {
  domain: string;
  policy: ACLPolicy;
  resources?: string[];
};

export type ACLDefaultRules = ACLRule[];
export type ACLGroupsRules = { [group: string]: ACLRule[]; };
export type ACLUsersRules = { [user: string]: ACLRule[]; };
export type ACLWhitelisted = { [domain: string]: string[]; };

export interface ACLConfiguration {
  default_policy?: ACLPolicy;
  whitelisted?: ACLWhitelisted;
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

  if (!newConfiguration.whitelisted) {
    newConfiguration.whitelisted = {};
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