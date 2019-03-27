
export type ACLPolicy = "deny" | "bypass" | "one_factor" | "two_factor";

export type ACLNetwork = string[];

export type ACLRule = {
  domain: string;
  resources?: string[];
  subject?: string;
  policy: ACLPolicy;
  networks?: ACLNetwork;
};

export interface ACLConfiguration {
  default_policy?: ACLPolicy;
  rules?: ACLRule[];
}

export function complete(configuration: ACLConfiguration): [ACLConfiguration, string[]] {
  const newConfiguration: ACLConfiguration = (configuration)
    ? JSON.parse(JSON.stringify(configuration)) : {};

  if (!newConfiguration.default_policy) {
    newConfiguration.default_policy = "bypass";
  }

  if (!newConfiguration.rules) {
    newConfiguration.rules = [];
  }

  if (newConfiguration.rules.length > 0) {
    const errors: string[] = [];
    newConfiguration.rules.forEach((r, idx) => {
      if (r.subject && !r.subject.match(/^(user|group):[a-zA-Z0-9]+$/)) {
        errors.push(`Rule ${idx} has wrong subject. It should be starting with user: or group:.`);
      }
    });
    if (errors.length > 0) {
      return [newConfiguration, errors];
    }
  }

  return [newConfiguration, []];
}