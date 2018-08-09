export interface SessionRedisOptions {
  host: string;
  port: number;
}

export interface SessionConfiguration {
  domain: string;
  secret: string;
  expiration?: number;
  inactivity?: number;
  redis?: SessionRedisOptions;
}

export function complete(configuration: SessionConfiguration): SessionConfiguration {
  const newConfiguration: SessionConfiguration = (configuration) ? JSON.parse(JSON.stringify(configuration)) : {};

  if (!newConfiguration.expiration) {
    newConfiguration.expiration = 3600000; // 1 hour
  }

  if (!newConfiguration.inactivity) {
    newConfiguration.inactivity = undefined; // disabled
  }

  return newConfiguration;
}