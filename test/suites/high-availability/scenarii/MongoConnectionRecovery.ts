import LoginAndRegisterTotp from "../../../helpers/LoginAndRegisterTotp";
import FullLogin from "../../../helpers/FullLogin";
import child_process from 'child_process';
import WithDriver from "../../../helpers/context/WithDriver";
import Logout from "../../../helpers/Logout";
import { composeFiles } from '../environment';
import DockerCompose from "../../../helpers/context/DockerCompose";

export default function() {
  after(async function() {
    await Logout(this.driver);
  })

  WithDriver();

  it("should be able to login after mongo restarts", async function() {
    this.timeout(30000);
    
    const secret = await LoginAndRegisterTotp(this.driver, "john", "password", true);
    const dockerCompose = new DockerCompose(composeFiles);
    await dockerCompose.restart('mongo');
    await FullLogin(this.driver, "john", secret, "https://admin.example.com:8080/secret.html");
  });  
}