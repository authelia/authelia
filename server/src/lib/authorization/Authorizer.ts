
import { ACLConfiguration, ACLPolicy, ACLRule } from "../configuration/schema/AclConfiguration";
import { IAuthorizer } from "./IAuthorizer";
import { Winston } from "../../../types/Dependencies";
import { MultipleDomainMatcher } from "./MultipleDomainMatcher";
import { Level } from "./Level";

function MatchDomain(actualDomain: string) {
  return function (rule: ACLRule): boolean {
    return MultipleDomainMatcher.match(actualDomain, rule.domain);
  };
}

function MatchResource(actualResource: string) {
  return function (rule: ACLRule): boolean {
    // If resources key is not provided, the rule applies to all resources.
    if (!rule.resources) return true;

    for (let i = 0; i < rule.resources.length; ++i) {
      const regexp = new RegExp(rule.resources[i]);
      if (regexp.test(actualResource)) return true;
    }
    return false;
  };
}

function MatchSubject(user: string, groups: string[]) {
  return (rule: ACLRule) => {
    // If no subject, matches anybody
    if (!rule.subject) return true;

    if (rule.subject.startsWith("user:")) {
      const ruleUser = rule.subject.split(":")[1];
      if (user == ruleUser) return true;
    }

    if (rule.subject.startsWith("group:")) {
      const ruleGroup = rule.subject.split(":")[1];
      if (groups.indexOf(ruleGroup) > -1) return true;
    }
    return false;
  };
}

export class Authorizer implements IAuthorizer {
  private logger: Winston;
  private readonly configuration: ACLConfiguration;

  constructor(configuration: ACLConfiguration, logger_: Winston) {
    this.logger = logger_;
    this.configuration = configuration;
  }

  private getMatchingRules(domain: string, resource: string, user: string, groups: string[]): ACLRule[] {
    const rules = this.configuration.rules;
    if (!rules) return [];
    return rules
      .filter(MatchDomain(domain))
      .filter(MatchResource(resource))
      .filter(MatchSubject(user, groups));
  }

  private ruleToLevel(policy: string): Level {
    if (policy == "bypass") {
      return Level.BYPASS;
    } else if (policy == "one_factor") {
      return Level.ONE_FACTOR;
    } else if (policy == "two_factor") {
      return Level.TWO_FACTOR;
    }
    return Level.DENY;
  }

  authorization(domain: string, resource: string, user: string, groups: string[]): Level {
    if (!this.configuration) return Level.BYPASS;

    const rules = this.getMatchingRules(domain, resource, user, groups);

    return (rules.length > 0)
      ? this.ruleToLevel(rules[0].policy) // extract the policy of the first matching rule
      : this.ruleToLevel(this.configuration.default_policy); // otherwise use the default policy
  }
}