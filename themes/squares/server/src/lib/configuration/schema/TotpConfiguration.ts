export interface TotpConfiguration {
  issuer: string;
}

export function complete(configuration: TotpConfiguration): TotpConfiguration {
  const newConfiguration: TotpConfiguration = (configuration) ? JSON.parse(JSON.stringify(configuration)) : {};

  if (!newConfiguration.issuer) {
    newConfiguration.issuer = "authelia.com";
  }

  return newConfiguration;
}