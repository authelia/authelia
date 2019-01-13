import { connect } from 'react-redux';
import OneTimePasswordRegistrationView, { OnSuccess, OnFailure } from '../../../views/OneTimePasswordRegistrationView/OneTimePasswordRegistrationView';
import { RootState } from '../../../reducers';
import { Dispatch } from 'redux';
import {to} from 'await-to-js';

const mapStateToProps = (state: RootState) => ({});

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

const mapDispatchToProps = (dispatch: Dispatch) => {
  return {
    componentDidMount: async (token: string, onSuccess: OnSuccess, onFailure: OnFailure) => {
      let err, result;
      [err, result] = await to(checkIdentity(token));
      if (err) {
        onFailure(err);
        return;
      }
      onSuccess(result.otpauth_url);
    }
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(OneTimePasswordRegistrationView);