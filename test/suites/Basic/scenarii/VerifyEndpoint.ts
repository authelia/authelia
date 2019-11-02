import { GET_Expect401, GET_ExpectRedirect } from "../../../helpers/utils/Requests";

export default function() {
  describe('Query without authenticated cookie', function() {
    it('should get a 401 on GET to http://192.168.240.1:9091/api/verify', async function() {
      await GET_Expect401('http://192.168.240.1:9091/api/verify', {
        'X-Forwarded-Proto': 'https',
      });
    });

    describe('Kubernetes nginx ingress controller', async function() {
      it('should redirect to https://login.example.com:8080', async function() {
        await GET_ExpectRedirect('http://192.168.240.1:9091/api/verify?rd=https://login.example.com:8080/%23/',
          'https://login.example.com:8080/#/?rd=https://secure.example.com:8080/',
          {
            'X-Original-Url': 'https://secure.example.com:8080/',
            'X-Forwarded-Proto': 'https'
          });
      });
    });

    describe('Traefik proxy', async function() {
      it('should redirect to https://login.example.com:8080', async function() {
        await GET_ExpectRedirect('http://192.168.240.1:9091/api/verify?rd=https://login.example.com:8080/%23/',
          'https://login.example.com:8080/#/?rd=https://secure.example.com:8080/',
          {
            'X-Forwarded-Proto': 'https',
            'X-Forwarded-Host': 'secure.example.com:8080',
            'X-Forwarded-Uri': '/',
          });
      });
    });
  });
}