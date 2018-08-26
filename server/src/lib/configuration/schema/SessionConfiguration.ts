export interface SessionRedisOptions {
  host: string;
  port: number;
  password?: string;
}

export interface SessionConfiguration {
  name?: string;
  domain: string;
  secret: string;
  expiration?: number;
  inactivity?: number;
  redis?: SessionRedisOptions;
}

export function complete(configuration: SessionConfiguration): SessionConfiguration {
  const newConfiguration: SessionConfiguration = (configuration) ? JSON.parse(JSON.stringify(configuration)) : {};

  if (!newConfiguration.name) {
    newConfiguration.name = "authelia_session";
  }

  if (!newConfiguration.expiration) {
    newConfiguration.expiration = 3600000; // 1 hour
  }

  if (!newConfiguration.inactivity) {
    newConfiguration.inactivity = undefined; // disabled
  }

  return newConfiguration;
}