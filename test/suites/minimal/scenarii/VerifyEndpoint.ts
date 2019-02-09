import { GET_Expect401, GET_ExpectRedirect } from "../../../helpers/utils/Requests";

export default function() {
  describe('Query without authenticated cookie', function() {
    it('should get a 401 on GET to https://authelia.example.com:8080/api/verify', async function() {
      await GET_Expect401('https://login.example.com:8080/api/verify');
    });

    describe('Parameter `rd` required by Kubernetes ingress controller', async function() {
      it('should redirect to https://login.example.com:8080', async function() {
        await GET_ExpectRedirect('https://login.example.com:8080/api/verify?rd=https://login.example.com:8080',
          'https://login.example.com:8080');
      });
    });
  });
}