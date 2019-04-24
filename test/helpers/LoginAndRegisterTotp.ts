import RegisterTotp from './RegisterTotp';
import LoginAs from './LoginAs';
import { WebDriver } from 'selenium-webdriver';
import VerifyIsSecondFactorStage from './assertions/VerifyIsSecondFactorStage';

export default async function(driver: WebDriver, user: string, password: string, email: boolean = false, timeout: number = 5000) {
  await LoginAs(driver, user, password, undefined, timeout);
  await VerifyIsSecondFactorStage(driver, timeout);
  return RegisterTotp(driver, email, timeout);
}