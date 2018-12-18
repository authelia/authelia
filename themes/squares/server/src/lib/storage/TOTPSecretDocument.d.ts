import { TOTPSecret } from "../../../types/TOTPSecret";

export interface TOTPSecretDocument {
  userid: string;
  secret: TOTPSecret;
}