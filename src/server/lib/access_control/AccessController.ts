
import { ACLConfiguration } from "../../../types/Configuration";
import PatternBuilder from "./PatternBuilder";
import { Winston } from "../../../types/Dependencies";

export class AccessController {
    private logger: Winston;
    private patternBuilder: PatternBuilder;

    constructor(configuration: ACLConfiguration, logger_: Winston) {
        this.logger = logger_;
        this.patternBuilder = new PatternBuilder(configuration, logger_);
    }

    isDomainAllowedForUser(domain: string, user: string, groups: string[]): boolean {
        const allowed_domains = this.patternBuilder.getAllowedDomains(user, groups);

        // Allow all matcher
        if (allowed_domains.length == 1 && allowed_domains[0] == "*") return true;

        this.logger.debug("ACL: trying to match %s with %s", domain,
            JSON.stringify(allowed_domains));
        for (let i = 0; i < allowed_domains.length; ++i) {
            const allowed_domain = allowed_domains[i];
            if (allowed_domain.startsWith("*") &&
                domain.endsWith(allowed_domain.substr(1))) {
                return true;
            }
            else if (domain == allowed_domain) {
                return true;
            }
        }
        return false;
    }
}