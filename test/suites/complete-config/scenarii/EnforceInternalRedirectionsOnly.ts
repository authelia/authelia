import LoginAndRegisterTotp from "../../../helpers/LoginAndRegisterTotp";
import VisitPage from "../../../helpers/VisitPage";
import FillLoginPageWithUserAndPasswordAndClick from '../../../helpers/FillLoginPageAndClick';
import ValidateTotp from "../../../helpers/ValidateTotp";
import Logout from "../../../helpers/Logout";
import WaitRedirected from "../../../helpers/WaitRedirected";
import IsAlreadyAuthenticatedStage from "../../../helpers/IsAlreadyAuthenticatedStage";

/*
 * Authelia should not be vulnerable to open redirection. Otherwise it would aid an
 * attacker in conducting a phishing attack.
 * 
 * To avoid the issue, Authelia's client scans the URL and prevent any redirection if
 * the URL is pointing to an external domain.
 */
export default function() {
  describe("Only redirection to a subdomain of the protected domain should be allowed", function() {
    this.timeout(10000);
    let secret: string;
  
    beforeEach(async function() {
      secret = await LoginAndRegisterTotp(this.driver, "john", true)
    });
  
    afterEach(async function() {
      await Logout(this.driver);
    })
  
    function CannotRedirectTo(url: string) {
      it(`should redirect to already authenticated page when requesting ${url}`, async function() {
        await VisitPage(this.driver, `https://login.example.com:8080/?rd=${url}`);
        await FillLoginPageWithUserAndPasswordAndClick(this.driver, 'john', 'password');
        await ValidateTotp(this.driver, secret);
        await IsAlreadyAuthenticatedStage(this.driver);
      });
    }

    function CanRedirectTo(url: string) {
      it(`should redirect to ${url}`, async function() {
        await VisitPage(this.driver, `https://login.example.com:8080/?rd=${url}`);
        await FillLoginPageWithUserAndPasswordAndClick(this.driver, 'john', 'password');
        await ValidateTotp(this.driver, secret);
        await WaitRedirected(this.driver, url);
      });
    }
    
    describe('blocked redirection', function() {
      // Do not redirect to another domain than example.com
      CannotRedirectTo("https://www.google.fr");

      // Do not redirect to rogue domain
      CannotRedirectTo("https://public.example.com.a:8080");

      // Do not redirect to http website
      CannotRedirectTo("http://public.example.com:8080");
    });

    describe('allowed redirection', function() {
      // Can redirect to any subdomain of the domain protected by Authelia.
      CanRedirectTo("https://public.example.com:8080/");
    });
  });
}