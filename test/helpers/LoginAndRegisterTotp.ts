import RegisterTotp from './RegisterTotp';
import LoginAs from './LoginAs';
import { WebDriver } from 'selenium-webdriver';
import IsSecondFactorStage from './IsSecondFactorStage';

export default async function(driver: WebDriver, user: string, email?: boolean) {
  await LoginAs(driver, user);
  await IsSecondFactorStage(driver);
  return await RegisterTotp(driver, email);
}