import LoginAndRegisterTotp from "../../../helpers/LoginAndRegisterTotp";
import FullLogin from "../../../helpers/FullLogin";
import WithDriver from "../../../helpers/context/WithDriver";
import Logout from "../../../helpers/Logout";
import { composeFiles } from '../environment';
import DockerCompose from "../../../helpers/context/DockerCompose";
import sleep from "../../../helpers/utils/sleep";

export default function () {
  const dockerCompose = new DockerCompose(composeFiles);

  WithDriver();

  it.only("should be able to login after mariadb restarts", async function () {
    this.timeout(30000);

    const secret = await LoginAndRegisterTotp(this.driver, "john", "password", true);
    await dockerCompose.restart('mariadb');
    await sleep(2000);

    await Logout(this.driver);
    await FullLogin(this.driver, "john", secret, "https://admin.example.com:8080/secret.html");
  });
}