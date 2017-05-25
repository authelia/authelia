
import { Winston } from "../../../types/Dependencies";
import { ACLConfiguration, ACLGroupsRules, ACLUsersRules, ACLDefaultRules } from "../../../types/Configuration";
import objectPath = require("object-path");

export default class AccessControlPatternBuilder {
    logger: Winston;
    configuration: ACLConfiguration;

    constructor(configuration: ACLConfiguration | undefined, logger_: Winston) {
        this.configuration = configuration;
        this.logger = logger_;
    }

    private buildFromGroups(groups: string[]): string[] {
        let allowed_domains: string[] = [];
        const groups_policy = objectPath.get<ACLConfiguration, ACLGroupsRules>(this.configuration, "groups");
        if (groups_policy) {
            for (let i = 0; i < groups.length; ++i) {
                const group = groups[i];
                if (group in groups_policy) {
                    const group_policy: string[] = groups_policy[group];
                    allowed_domains = allowed_domains.concat(groups_policy[group]);
                }
            }
        }
        return allowed_domains;
    }

    private buildFromUser(user: string): string[] {
        let allowed_domains: string[] = [];
        const users_policy = objectPath.get<ACLConfiguration, ACLUsersRules>(this.configuration, "users");
        if (users_policy) {
            if (user in users_policy) {
                allowed_domains = allowed_domains.concat(users_policy[user]);
            }
        }
        return allowed_domains;
    }

    getAllowedDomains(user: string, groups: string[]): string[] {
        if (!this.configuration) {
            this.logger.debug("No access control rules found." +
                "Default policy to allow all.");
            return ["*"]; // No configuration means, no restrictions.
        }

        let allowed_domains: string[] = [];
        const default_policy = objectPath.get<ACLConfiguration, ACLDefaultRules>(this.configuration, "default");
        if (default_policy) {
            allowed_domains = allowed_domains.concat(default_policy);
        }

        allowed_domains = allowed_domains.concat(this.buildFromGroups(groups));
        allowed_domains = allowed_domains.concat(this.buildFromUser(user));

        this.logger.debug("ACL: user \'%s\' is allowed access to %s", user,
            JSON.stringify(allowed_domains));
        return allowed_domains;
    }
}
