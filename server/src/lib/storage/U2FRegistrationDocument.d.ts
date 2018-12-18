
import { U2FRegistration } from "../../../types/U2FRegistration";

export interface U2FRegistrationDocument {
  userId: string;
  appId: string;
  registration: U2FRegistration;
}