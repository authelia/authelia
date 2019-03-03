import { connect } from 'react-redux';
import OneTimePasswordRegistrationView from '../../../views/OneTimePasswordRegistrationView/OneTimePasswordRegistrationView';
import { RootState } from '../../../reducers';
import { Dispatch } from 'redux';
import {to} from 'await-to-js';
import { generateTotpSecret, generateTotpSecretSuccess, generateTotpSecretFailure } from '../../../reducers/Portal/OneTimePasswordRegistration/actions';
import { push } from 'connected-react-router';

const mapStateToProps = (state: RootState) => ({
  error: state.oneTimePasswordRegistration.error,
  secret: state.oneTimePasswordRegistration.secret,
});

async function checkIdentity(token: string) {
  return fetch(`/api/secondfactor/totp/identity/finish?token=${token}`, {
    method: 'POST',
    headers: {
      'Accept': 'application/json',
      'Content-Type': 'application/json',
    },
  })
    .then(async (res) => {
      if (res.status !== 200) {
        throw new Error('Status code ' + res.status);
      }

      const body = await res.json();
      if ('error' in body) {
        throw new Error(body['error']);
      }
      return body;
    });
}

async function tryGenerateTotpSecret(dispatch: Dispatch, token: string) {
  let err, result;
  dispatch(generateTotpSecret());
  [err, result] = await to(checkIdentity(token));
  if (err) {
    const e = err;
    setTimeout(() => {
      dispatch(generateTotpSecretFailure(e.message));
    }, 2000);
    return;
  }
  dispatch(generateTotpSecretSuccess(result));
}

const mapDispatchToProps = (dispatch: Dispatch) => {
  let internalToken: string;
  return {
    onInit: async (token: string) => {
      internalToken = token;
      await tryGenerateTotpSecret(dispatch, internalToken);
    },
    onRetryClicked: async () => {
      await tryGenerateTotpSecret(dispatch, internalToken);
    },
    onCancelClicked: () => {
      dispatch(push('/'));
    },
    onLoginClicked: () => {
      dispatch(push('/'));
    }
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(OneTimePasswordRegistrationView);