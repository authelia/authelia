import { TOTPConfiguration } from "../Configuration";
import { ObjectCloner } from "../../utils/ObjectCloner";

const DEFAULT_ISSUER = "authelia.com";

export class TOTPAdapter {
  static adapt(configuration: TOTPConfiguration): TOTPConfiguration {
    const newConfiguration = {
      issuer: DEFAULT_ISSUER
    };

    if (!configuration)
      return newConfiguration;

    if (configuration && configuration.issuer)
      newConfiguration.issuer = configuration.issuer;

    return newConfiguration;
  }
}