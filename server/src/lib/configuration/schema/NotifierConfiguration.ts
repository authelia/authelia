
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

export function complete(configuration: NotifierConfiguration): [NotifierConfiguration, string] {
  const newConfiguration: NotifierConfiguration = (configuration) ? JSON.parse(JSON.stringify(configuration)) : {};

  if (Object.keys(newConfiguration).length == 0)
    newConfiguration.filesystem = { filename: "/tmp/authelia/notification.txt" };

  const ERROR = "Notifier must have one of the following keys: 'filesystem', 'email' or 'smtp'";

  if (Object.keys(newConfiguration).length != 1)
    return [newConfiguration, ERROR];

  const key = Object.keys(newConfiguration)[0];

  if (key != "filesystem" && key != "smtp" && key != "email")
    return [newConfiguration, ERROR];

  return [newConfiguration, undefined];
}
