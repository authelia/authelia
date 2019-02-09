import { POST_Expect401, GET_Expect401 } from "../../../helpers/utils/Requests";

export default function() {
  // POST
  it('should return 401 error when posting to https://login.example.com:8080/api/totp', async function() {
    await POST_Expect401('https://login.example.com:8080/api/totp', { token: 'MALICIOUS_TOKEN' });
  });

  it('should return 401 error when posting to https://login.example.com:8080/api/u2f/sign', async function() {
    await POST_Expect401('https://login.example.com:8080/api/u2f/sign');
  });

  it('should return 401 error when posting to https://login.example.com:8080/api/u2f/register', async function() {
    await POST_Expect401('https://login.example.com:8080/api/u2f/register');
  });

  
  // GET
  it('should return 401 error on GET to https://login.example.com:8080/api/u2f/sign_request', async function() {
    await GET_Expect401('https://login.example.com:8080/api/u2f/sign_request');
  });

  it('should return 401 error on GET to https://login.example.com:8080/api/u2f/register_request', async function() {
    await GET_Expect401('https://login.example.com:8080/api/u2f/register_request');
  });


  describe('Identity validation endpoints blocked to unauthenticated users', function() {
    it('should return 401 error on POST to https://login.example.com:8080/api/secondfactor/u2f/identity/start', async function() {
      await POST_Expect401('https://login.example.com:8080/api/secondfactor/u2f/identity/start');
    });

    it('should return 401 error on POST to https://login.example.com:8080/api/secondfactor/u2f/identity/finish', async function() {
      await POST_Expect401('https://login.example.com:8080/api/secondfactor/u2f/identity/finish');
    });
  
    it('should return 401 error on POST to https://login.example.com:8080/api/secondfactor/totp/identity/start', async function() {
      await POST_Expect401('https://login.example.com:8080/api/secondfactor/totp/identity/start');
    });

    it('should return 401 error on POST to https://login.example.com:8080/api/secondfactor/totp/identity/finish', async function() {
      await POST_Expect401('https://login.example.com:8080/api/secondfactor/totp/identity/finish');
    });
  });
}