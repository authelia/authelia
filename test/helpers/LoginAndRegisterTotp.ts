import RegisterTotp from './RegisterTotp';
import LoginAs from './LoginAs';
import { WebDriver } from 'selenium-webdriver';
import VerifyIsSecondFactorStage from './assertions/VerifyIsSecondFactorStage';

export default async function(driver: WebDriver, user: string, password: string, email: boolean = false) {
  await LoginAs(driver, user, password);
  await VerifyIsSecondFactorStage(driver);
  return await RegisterTotp(driver, email);
}