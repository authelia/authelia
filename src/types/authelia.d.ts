
import * as winston from "winston";
import * as nedb from "nedb";

declare namespace authelia {

    interface LdapConfiguration {
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
    type ACLGroupsRules = Map<GroupName, DomainPattern>;
    type ACLUsersRules = Map<UserName, DomainPattern>;

    export interface ACLConfiguration {
        default: ACLDefaultRules;
        groups: ACLGroupsRules;
        users: ACLUsersRules;
    }

    interface SessionCookieConfiguration {
        secret: string;
        expiration: number;
        domain: string
    }

    type NotifierType = string;
    export type NotifiersConfiguration = Map<NotifierType, any>;

    export interface Configuration {
        port: number;
        logs_level: string;
        ldap: LdapConfiguration | {};
        session_domain?: string;
        session_secret: string;
        session_max_age: number;
        store_directory?: string;
        notifier: NotifiersConfiguration;
        access_control: ACLConfiguration;
    }

    export interface GlobalDependencies {
        u2f: object;
        nodemailer: any;
        ldapjs: object;
        session: any;
        winston: winston.Winston;
        speakeasy: object;
        nedb: object;
    }
}