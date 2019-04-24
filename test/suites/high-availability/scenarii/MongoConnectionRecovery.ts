import LoginAndRegisterTotp from "../../../helpers/LoginAndRegisterTotp";
import FullLogin from "../../../helpers/FullLogin";
import WithDriver from "../../../helpers/context/WithDriver";
import Logout from "../../../helpers/Logout";
import { composeFiles } from '../environment';
import DockerCompose from "../../../helpers/context/DockerCompose";

export default function() {
  const dockerCompose = new DockerCompose(composeFiles);

  WithDriver();

  it("should be able to login after mongo restarts", async function() {
    this.timeout(30000);
    
    const secret = await LoginAndRegisterTotp(this.driver, "john", "password", true);
    await dockerCompose.restart('mongo');

    await Logout(this.driver);
    await FullLogin(this.driver, "john", secret, "https://admin.example.com:8080/secret.html");
    // TODO(clems4ever): logout here but right now visiting login.example.com redirects to home.example.com
    // according to the configuration so it's not possible to click on Logout link.
  });
}