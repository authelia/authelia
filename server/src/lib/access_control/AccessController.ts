
import { ACLConfiguration, ACLPolicy, ACLRule, ACLWhitelisted } from "../configuration/schema/AclConfiguration";
import { IAccessController } from "./IAccessController";
import { Winston } from "../../../types/Dependencies";
import { MultipleDomainMatcher } from "./MultipleDomainMatcher";

import ipRangeCheck = require("ip-range-check");


enum AccessReturn {
  NO_MATCHING_RULES,
  MATCHING_RULES_AND_ACCESS,
  MATCHING_RULES_AND_NO_ACCESS
}

function AllowedRule(rule: ACLRule) {
  return rule.policy == "allow";
}

function MatchDomain(actualDomain: string) {
  return function (rule: ACLRule): boolean {
    return MultipleDomainMatcher.match(actualDomain, rule.domain);
  };
}

function MatchWhitelistDomain(actualDomain: string) {
    return function (rule: string): boolean {
        return MultipleDomainMatcher.match(actualDomain, rule);
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

function SelectPolicy(rule: ACLRule): ("allow" | "deny") {
  return rule.policy;
}

export class AccessController implements IAccessController {
  private logger: Winston;
  private readonly configuration: ACLConfiguration;

  constructor(configuration: ACLConfiguration, logger_: Winston) {
    this.logger = logger_;
    this.configuration = configuration;
  }

  private isAccessAllowedInRules(rules: ACLRule[], domain: string, resource: string): AccessReturn {
    if (!rules)
      return AccessReturn.NO_MATCHING_RULES;

    const policies = rules.map(SelectPolicy);

    if (rules.length > 0) {
      if (policies[0] == "allow") {
        return AccessReturn.MATCHING_RULES_AND_ACCESS;
      }
      else {
        return AccessReturn.MATCHING_RULES_AND_NO_ACCESS;
      }
    }
    return AccessReturn.NO_MATCHING_RULES;
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

  private getMatchingWhitelistRules(domain: string): string[][] {
    const rules = this.configuration.whitelisted;
    if (!rules) return [];
    return Object.keys(rules).filter(MatchWhitelistDomain(domain)).map(key => rules[key]);
  }

  private isAccessAllowedDefaultPolicy(): boolean {
    return this.configuration.default_policy == "allow";
  }

  isWhitelisted(domain: string, ip: string): boolean {
    const rules = this.getMatchingWhitelistRules(domain);
    if (rules.length > 0) {
      for (let i = 0; i < rules.length; ++i) {
        if (ipRangeCheck(ip, rules[i])) {
          return true;
        }
      }
    }
    return false;
  }

  isAccessAllowed(domain: string, resource: string, user: string, groups: string[]): boolean {
    if (!this.configuration) return true;

    const allRules = this.getMatchingAllRules(domain, resource);
    const groupRules = this.getMatchingGroupRules(groups, domain, resource);
    const userRules = this.getMatchingUserRules(user, domain, resource);
    const rules = allRules.concat(groupRules).concat(userRules).reverse();

    const access = this.isAccessAllowedInRules(rules, domain, resource);
    if (access == AccessReturn.MATCHING_RULES_AND_ACCESS)
      return true;
    else if (access == AccessReturn.MATCHING_RULES_AND_NO_ACCESS)
      return false;

    return this.isAccessAllowedDefaultPolicy();
  }
}