import { POST_Expect403, GET_Expect403 } from "../../../helpers/utils/Requests";

export default function() {
  // POST
  it('should return 403 error when posting to https://login.example.com:8080/api/secondfactor/totp', async function() {
    await POST_Expect403('https://login.example.com:8080/api/secondfactor/totp', { token: 'MALICIOUS_TOKEN' });
  });

  it('should return 403 error when posting to https://login.example.com:8080/api/secondfactor/u2f/sign', async function() {
    await POST_Expect403('https://login.example.com:8080/api/secondfactor/u2f/sign');
  });

  it('should return 403 error when posting to https://login.example.com:8080/api/secondfactor/u2f/register', async function() {
    await POST_Expect403('https://login.example.com:8080/api/secondfactor/u2f/register');
  });
  
  it('should return 403 error on GET to https://login.example.com:8080/api/secondfactor/u2f/sign_request', async function() {
    await POST_Expect403('https://login.example.com:8080/api/secondfactor/u2f/sign_request');
  });

  it('should return 403 error when posting to https://login.example.com:8080/api/secondfactor/preferences', async function() {
    await POST_Expect403('https://login.example.com:8080/api/secondfactor/preferences');
  });

  it('should return 403 error on GET to https://login.example.com:8080/api/secondfactor/preferences', async function() {
    await GET_Expect403('https://login.example.com:8080/api/secondfactor/preferences');
  });
  
  it('should return 403 error on GET to https://login.example.com:8080/api/secondfactor/available', async function() {
    await GET_Expect403('https://login.example.com:8080/api/secondfactor/available');
  });


  describe('Identity validation endpoints blocked to unauthenticated users', function() {
    it('should return 403 error on POST to https://login.example.com:8080/api/secondfactor/u2f/identity/start', async function() {
      await POST_Expect403('https://login.example.com:8080/api/secondfactor/u2f/identity/start');
    });

    it('should return 403 error on POST to https://login.example.com:8080/api/secondfactor/u2f/identity/finish', async function() {
      await POST_Expect403('https://login.example.com:8080/api/secondfactor/u2f/identity/finish');
    });
  
    it('should return 403 error on POST to https://login.example.com:8080/api/secondfactor/totp/identity/start', async function() {
      await POST_Expect403('https://login.example.com:8080/api/secondfactor/totp/identity/start');
    });

    it('should return 403 error on POST to https://login.example.com:8080/api/secondfactor/totp/identity/finish', async function() {
      await POST_Expect403('https://login.example.com:8080/api/secondfactor/totp/identity/finish');
    });
  });
}