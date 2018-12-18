export interface RegulationConfiguration {
  max_retries?: number;
  find_time?: number;
  ban_time?: number;
}

export function complete(configuration: RegulationConfiguration): RegulationConfiguration {
  const newConfiguration: RegulationConfiguration = (configuration) ? JSON.parse(JSON.stringify(configuration)) : {};

  if (!newConfiguration.max_retries) {
    newConfiguration.max_retries = 3;
  }

  if (!newConfiguration.find_time) {
    newConfiguration.find_time = 120; // seconds
  }

  if (!newConfiguration.ban_time) {
    newConfiguration.ban_time = 300; // seconds
  }

  return newConfiguration;
}