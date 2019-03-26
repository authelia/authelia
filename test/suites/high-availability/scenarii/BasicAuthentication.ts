import Request from 'request-promise';

async function GetSecret(username: string, password: string) {
  return await Request('https://singlefactor.example.com:8080/secret.html', {
    auth: {
      username,
      password
    },
    rejectUnauthorized: false,
  });
}

export default function() {
  it("should retrieve secret when Proxy-Authorization header is provided", async function() {
    const res = await GetSecret('john', 'password');
    if (res.indexOf('This is a very important secret!') < 0) {
      throw new Error('Cannot access secret.');
    }
  });

  it("should not retrieve secret when providing bad password", async function() {
    const res = await GetSecret('john', 'bad-password');
    if (res.indexOf('This is a very important secret!') >= 0) {
      throw new Error('Cannot access secret.');
    }
  });

  it("should not retrieve secret when authenticating with unexisting user", async function() {
    const res = await GetSecret('dontexist', 'password');
    if (res.indexOf('This is a very important secret!') >= 0) {
      throw new Error('Cannot access secret.');
    }
  });
}