
import { ACLConfiguration, ACLPolicy, ACLRule } from "../configuration/schema/AclConfiguration";
import { IAccessController } from "./IAccessController";
import { Winston } from "../../../types/Dependencies";
import { MultipleDomainMatcher } from "./MultipleDomainMatcher";


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

function MatchWhitelist(whitelisted: boolean) {
  return function (rule: ACLRule): boolean {
    // If not a whitelisted user ignore this filter
    if (!whitelisted) return true;

    // If not whitelist_policy defined for the user, fall back to the default_whitelist_policy
    if (!rule.whitelist_policy) return false;

    return rule.whitelist_policy === "allow";
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

  private getMatchingUserRules(user: string, domain: string, resource: string, whitelisted: boolean): ACLRule[] {
    const userRules = this.configuration.users[user];
    if (!userRules) return [];
    return userRules.filter(MatchDomain(domain)).filter(MatchResource(resource)).filter(MatchWhitelist(whitelisted));
  }

  private getMatchingGroupRules(groups: string[], domain: string, resource: string, whitelisted: boolean): ACLRule[] {
    const that = this;
    // There is no ordering between group rules. That is, when a user belongs to 2 groups, there is no
    // guarantee one set of rules has precedence on the other one.
    const groupRules = groups.reduce(function (rules: ACLRule[], group: string) {
      const groupRules = that.configuration.groups[group];
      if (groupRules) rules = rules.concat(groupRules);
      return rules;
    }, []);
    return groupRules.filter(MatchDomain(domain)).filter(MatchResource(resource)).filter(MatchWhitelist(whitelisted));
  }

  private getMatchingAllRules(domain: string, resource: string, whitelisted: boolean): ACLRule[] {
    const rules = this.configuration.any;
    if (!rules) return [];
    return rules.filter(MatchDomain(domain)).filter(MatchResource(resource)).filter(MatchWhitelist(whitelisted));
  }

  private isAccessAllowedDefaultPolicy(): boolean {
    return this.configuration.default_policy == "allow";
  }

  private isAccessAllowedDefaultWhitelistPolicy(): boolean {
    return this.configuration.default_whitelist_policy == "allow";
  }

  isAccessAllowed(domain: string, resource: string, user: string, groups: string[], whitelisted: boolean): boolean {
    if (!this.configuration) return true;

    const allRules = this.getMatchingAllRules(domain, resource, whitelisted);
    const groupRules = this.getMatchingGroupRules(groups, domain, resource, whitelisted);
    const userRules = this.getMatchingUserRules(user, domain, resource, whitelisted);
    const rules = allRules.concat(groupRules).concat(userRules).reverse();

    const access = this.isAccessAllowedInRules(rules, domain, resource);
    if (access == AccessReturn.MATCHING_RULES_AND_ACCESS)
      return true;
    else if (access == AccessReturn.MATCHING_RULES_AND_NO_ACCESS)
      return false;

    if (whitelisted) {
      return this.isAccessAllowedDefaultWhitelistPolicy();
    }

    return this.isAccessAllowedDefaultPolicy();
  }
}