
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

export class Authorizer implements IAuthorizer {
  private logger: Winston;
  private readonly configuration: ACLConfiguration;

  constructor(configuration: ACLConfiguration, logger_: Winston) {
    this.logger = logger_;
    this.configuration = configuration;
  }

  private getMatchingUserRules(user: string, domain: string, resource: string): ACLRule[] {
    const userRules = this.configuration.users[user];
    if (!userRules) return [];
    return userRules.filter(MatchDomain(domain)).filter(MatchResource(resource));
  }

  private getMatchingGroupRules(groups: string[], domain: string, resource: string): ACLRule[] {
    const that = this;
    // There is no ordering between group rules. That is, when a user belongs to 2 groups, there is no
    // guarantee one set of rules has precedence on the other one.
    const groupRules = groups.reduce(function (rules: ACLRule[], group: string) {
      const groupRules = that.configuration.groups[group];
      if (groupRules) rules = rules.concat(groupRules);
      return rules;
    }, []);
    return groupRules.filter(MatchDomain(domain)).filter(MatchResource(resource));
  }

  private getMatchingAllRules(domain: string, resource: string): ACLRule[] {
    const rules = this.configuration.any;
    if (!rules) return [];
    return rules.filter(MatchDomain(domain)).filter(MatchResource(resource));
  }

  authorization(domain: string, resource: string, user: string, groups: string[]): Level {
    if (!this.configuration) return Level.BYPASS;

    const allRules = this.getMatchingAllRules(domain, resource);
    const groupRules = this.getMatchingGroupRules(groups, domain, resource);
    const userRules = this.getMatchingUserRules(user, domain, resource);
    const rules = allRules.concat(groupRules).concat(userRules).reverse();
    const policy = rules.map(r => r.policy).concat([this.configuration.default_policy])[0];

    if (policy == "bypass") {
      return Level.BYPASS;
    } else if (policy == "one_factor") {
      return Level.ONE_FACTOR;
    } else if (policy == "two_factor") {
      return Level.TWO_FACTOR;
    }
    return Level.DENY;
  }
}