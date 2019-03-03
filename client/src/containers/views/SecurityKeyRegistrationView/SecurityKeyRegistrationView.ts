import { connect } from 'react-redux';
import SecurityKeyRegistrationView from '../../../views/SecurityKeyRegistrationView/SecurityKeyRegistrationView';
import { RootState } from '../../../reducers';
import { Dispatch } from 'redux';
import {to} from 'await-to-js';
import * as U2fApi from "u2f-api";
import { Props } from '../../../views/SecurityKeyRegistrationView/SecurityKeyRegistrationView';
import { registerSecurityKey, registerSecurityKeyFailure, registerSecurityKeySuccess } from '../../../reducers/Portal/SecurityKeyRegistration/actions';

const mapStateToProps = (state: RootState) => ({
  deviceRegistered: state.securityKeyRegistration.success,
  error: state.securityKeyRegistration.error,
});

async function checkIdentity(token: string) {
  return fetch(`/api/secondfactor/u2f/identity/finish?token=${token}`, {
    method: 'POST',
  });
}

async function requestRegistration() {
  return fetch('/api/u2f/register_request')
    .then(async (res) => {
      if (res.status !== 200) {
        throw new Error('Status code ' + res.status);
      }
      return res.json();
    });
}

async function completeRegistration(response: U2fApi.RegisterResponse) {
  return fetch('/api/u2f/register', {
    method: 'POST',
    headers: {
      'Accept': 'application/json',
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(response),
  })
    .then(async (res) => {
      if (res.status !== 200) {
        throw new Error('Status code ' + res.status);
      }
    });
}

function fail(dispatch: Dispatch, err: Error) {
  dispatch(registerSecurityKeyFailure(err.message));
}

const mapDispatchToProps = (dispatch: Dispatch, ownProps: Props) => {
  return {
    onInit: async (token: string) => {
      let err, result;
      dispatch(registerSecurityKey());
      [err, result] = await to(checkIdentity(token));
      if (err) {
        fail(dispatch, err);
        return;
      }
      [err, result] = await to(requestRegistration());
      if (err) {
        fail(dispatch, err);
        return;
      }

      [err, result] = await to(U2fApi.register(result, [], 60));
      if (err) {
        fail(dispatch, err);
        return;
      }

      [err, result] = await to(completeRegistration(result as U2fApi.RegisterResponse));
      if (err) {
        fail(dispatch, err);
        return;
      }

      dispatch(registerSecurityKeySuccess());
      setTimeout(() => {
        ownProps.history.push('/2fa');
      }, 2000);
    },
    onBackClicked: () => {
      ownProps.history.push('/2fa');
    }
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(SecurityKeyRegistrationView);