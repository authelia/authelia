export interface UserLdapConfiguration {
  url: string;
  base_dn: string;

  additional_users_dn?: string;
  users_filter?: string;

  additional_groups_dn?: string;
  groups_filter?: string;

  group_name_attribute?: string;
  mail_attribute?: string;

  user: string; // admin username
  password: string; // admin password
}

export interface LdapConfiguration {
  url: string;

  users_dn: string;
  users_filter: string;

  groups_dn: string;
  groups_filter: string;

  group_name_attribute: string;
  mail_attribute: string;

  user: string; // admin username
  password: string; // admin password
}

type UserName = string;
type GroupName = string;
type DomainPattern = string;

export type ACLPolicy = 'deny' | 'allow';

export type ACLRule = {
  domain: string;
  policy: ACLPolicy;
  resources?: string[];
}

export type ACLDefaultRules = ACLRule[];
export type ACLGroupsRules = { [group: string]: ACLRule[]; };
export type ACLUsersRules = { [user: string]: ACLRule[]; };

export interface ACLConfiguration {
  default_policy?: ACLPolicy;
  any?: ACLDefaultRules;
  groups?: ACLGroupsRules;
  users?: ACLUsersRules;
}

export interface SessionRedisOptions {
  host: string;
  port: number;
}

interface SessionCookieConfiguration {
  secret: string;
  expiration?: number;
  inactivity?: number;
  domain?: string;
  redis?: SessionRedisOptions;
}

export interface EmailNotifierConfiguration {
  username: string;
  password: string;
  sender: string;
  service: string;
}

export interface SmtpNotifierConfiguration {
  username?: string;
  password?: string;
  host: string;
  port: number;
  secure: boolean;
  sender: string;
}

export interface FileSystemNotifierConfiguration {
  filename: string;
}

export interface NotifierConfiguration {
  email?: EmailNotifierConfiguration;
  smtp?: SmtpNotifierConfiguration;
  filesystem?: FileSystemNotifierConfiguration;
}

export interface MongoStorageConfiguration {
  url: string;
  database: string;
}

export interface LocalStorageConfiguration {
  path?: string;
  in_memory?: boolean;
}

export interface StorageConfiguration {
  local?: LocalStorageConfiguration;
  mongo?: MongoStorageConfiguration;
}

export interface RegulationConfiguration {
  max_retries: number;
  find_time: number;
  ban_time: number;
}

declare type AuthenticationMethod = 'two_factor' | 'single_factor';
declare type AuthenticationMethodPerSubdomain = { [subdomain: string]: AuthenticationMethod }

export interface AuthenticationMethodsConfiguration {
  default_method: AuthenticationMethod;
  per_subdomain_methods?: AuthenticationMethodPerSubdomain;
}

export interface TOTPConfiguration {
  issuer: string;
}

export interface UserConfiguration {
  port?: number;
  logs_level?: string;
  ldap: UserLdapConfiguration;
  session: SessionCookieConfiguration;
  storage: StorageConfiguration;
  notifier: NotifierConfiguration;
  authentication_methods?: AuthenticationMethodsConfiguration;
  access_control?: ACLConfiguration;
  regulation: RegulationConfiguration;
  default_redirection_url?: string;
  totp?: TOTPConfiguration;
}

export interface AppConfiguration {
  port: number;
  logs_level: string;
  ldap: LdapConfiguration;
  session: SessionCookieConfiguration;
  storage: StorageConfiguration;
  notifier: NotifierConfiguration;
  authentication_methods: AuthenticationMethodsConfiguration;
  access_control?: ACLConfiguration;
  regulation: RegulationConfiguration;
  default_redirection_url?: string;
  totp: TOTPConfiguration;
}
