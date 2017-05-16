
export interface LdapConfiguration {
    url: string;
    base_dn: string;
    additional_user_dn?: string;
    user_name_attribute?: string; // cn by default
    additional_group_dn?: string;
    group_name_attribute?: string; // cn by default
    user: string; // admin username
    password: string; // admin password
}

type UserName = string;
type GroupName = string;
type DomainPattern = string;

type ACLDefaultRules = Array<DomainPattern>;
type ACLGroupsRules = Object;
type ACLUsersRules = Object;

export interface ACLConfiguration {
    default: ACLDefaultRules;
    groups: ACLGroupsRules;
    users: ACLUsersRules;
}

interface SessionCookieConfiguration {
    secret: string;
    expiration?: number;
    domain?: string;
}

interface GMailNotifier {
    user: string;
    pass: string;
}

type NotifierType = string;
export interface NotifiersConfiguration {
    gmail: GMailNotifier;
}

export interface UserConfiguration {
    port?: number;
    logs_level?: string;
    ldap: LdapConfiguration;
    session: SessionCookieConfiguration;
    store_directory?: string;
    notifier: NotifiersConfiguration;
    access_control?: ACLConfiguration;
}

export interface AppConfiguration {
    port: number;
    logs_level: string;
    ldap: LdapConfiguration;
    session: SessionCookieConfiguration;
    store_in_memory?: boolean;
    store_directory?: string;
    notifier: NotifiersConfiguration;
    access_control?: ACLConfiguration;
}
