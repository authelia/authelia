export interface MongoStorageConfiguration {
  url: string;
  database: string;
  auth?: {
    username: string;
    password: string;
  };
}

export interface LocalStorageConfiguration {
  path?: string;
  in_memory?: boolean;
}

export interface StorageConfiguration {
  local?: LocalStorageConfiguration;
  mongo?: MongoStorageConfiguration;
}

export function complete(configuration: StorageConfiguration): StorageConfiguration {
  const newConfiguration: StorageConfiguration = (configuration) ? JSON.parse(JSON.stringify(configuration)) : {};

  if (!newConfiguration.local && !newConfiguration.mongo) {
    newConfiguration.local = {
      in_memory: true
    };
  }

  return newConfiguration;
}