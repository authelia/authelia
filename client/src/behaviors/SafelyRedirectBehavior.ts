import { Dispatch } from "redux";
import * as AutheliaService from '../services/AutheliaService';

export default async function(url: string) {
  try {
    // Check the url against the backend before redirecting.
    await AutheliaService.checkRedirection(url);
    window.location.href = url;
  } catch (e) {
    console.error(
      'Cannot redirect since the URL is not in the protected domain.' +
      'This behavior could be malicious so please the issue to an administrator.');
    throw e;
  }
}