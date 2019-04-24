import Request from 'request-promise';
import Fetch from 'node-fetch';
import Assert from 'assert';
import { StatusCodeError } from 'request-promise/errors';

process.env.NODE_TLS_REJECT_UNAUTHORIZED = "0";

export async function GET_ExpectError(url: string, headers: {[key: string]: string}, statusCode: number) {
  try {
    await Request.get(url, {
      json: true,
      rejectUnauthorized: false,
      headers: headers,
    });
    throw new Error('No response');
  } catch (e) {
    if (e instanceof StatusCodeError) {
      Assert.equal(e.statusCode, statusCode);
      return;
    }
  }
  return;
}

// Sent a GET request to the url and expect a 401
export async function GET_Expect401(url: string, headers: {[key: string]: string} = {}) {
  return await GET_ExpectError(url, headers, 401);
}

export async function GET_Expect403(url: string, headers: {[key: string]: string} = {}) {
  return await GET_ExpectError(url, headers, 403);
}

export async function GET_Expect502(url: string, headers: {[key: string]: string} = {}) {
  return await GET_ExpectError(url, headers, 502);
}

export async function POST_Expect403(url: string, body?: any) {
  try {
    await Request.post(url, {
      json: true,
      rejectUnauthorized: false,
      body
    });
    throw new Error('No response');
  } catch (e) {
    if (e instanceof StatusCodeError) {
      Assert.equal(e.statusCode, 403);
      return;
    }
  }
  return;
}

export async function GET_ExpectRedirect(url: string, redirectionUrl: string, headers: {[key: string]: string} = {}) {
  const response = await Fetch(url, {redirect: 'manual', headers: headers});

  if (response.status == 302) {
    const body = await response.text();
    Assert.equal(body, 'Found. Redirecting to ' + redirectionUrl);
    return;
  }

  throw new Error('No redirect');
}