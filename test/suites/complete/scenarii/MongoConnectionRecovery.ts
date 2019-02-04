import LoginAndRegisterTotp from "../../../helpers/LoginAndRegisterTotp";
import FullLogin from "../../../helpers/FullLogin";
import child_process from 'child_process';

export default function() {
  it("should be able to login after mongo restarts", async function() {
    this.timeout(30000);
    
    const secret = await LoginAndRegisterTotp(this.driver, "john", true);
    child_process.execSync("./scripts/dc-dev.sh restart mongo");
    await FullLogin(this.driver, "https://admin.example.com:8080/secret.html", "john", secret);
  });  
}