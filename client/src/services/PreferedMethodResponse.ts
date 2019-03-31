import Method2FA from "../types/Method2FA";


export default interface PreferedMethodResponse {
  method?: Method2FA;
  error?: string;
}