import { connect } from 'react-redux';
import SecondFactorView, {Props} from '../../../views/SecondFactorView/SecondFactorView';
import { RootState } from '../../../reducers';
import { Dispatch } from 'redux';
import u2fApi, { SignResponse } from 'u2f-api';
import to from 'await-to-js';
import { logoutSuccess, logoutFailure, logout } from '../../../reducers/Portal/actions';
import AuthenticationLevel from '../../../types/AuthenticationLevel';
import RemoteState from '../../../reducers/Portal/RemoteState';

const mapStateToProps = (state: RootState) => ({
  state: state.remoteState,
  stateError: state.remoteStateError,
});

async function requestSigning() {
  return fetch('/api/u2f/sign_request')
    .then(async (res) => {
      if (res.status !== 200) {
        throw new Error('Status code ' + res.status);
      }
      return res.json();
    });
}

async function completeSigning(response: u2fApi.SignResponse) {
  return fetch('/api/u2f/sign', {
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

async function triggerSecurityKeySigning() {
  let err, result;
  [err, result] = await to(requestSigning());
  if (err) {
    console.error(err);
    return;
  }

  [err, result] = await to(u2fApi.sign(result, 60));
  if (err) {
    console.error(err);
    return;
  }

  [err, result] = await to(completeSigning(result as SignResponse));
  if (err) {
    console.error(err);
    return;
  }
}

const mapDispatchToProps = (dispatch: Dispatch, ownProps: Props) => {
  return {
    onLogoutClicked: () => {
      dispatch(logout());
      fetch('/api/logout', {
        method: 'POST',
      })
        .then(async (res) => {
          if (res.status != 200) {
            throw new Error('Status code ' + res.status);
          }
          await dispatch(logoutSuccess());
          ownProps.history.push('/');
        })
        .catch(async (err: string) => {
          console.error(err);
          await dispatch(logoutFailure(err));
        });
    },
    onRegisterSecurityKeyClicked: () => {
      fetch('/api/secondfactor/u2f/identity/start', {
        method: 'POST',
      })
      .then(async (res) => {
        if (res.status != 200) {
          throw new Error('Status code ' + res.status);
        }
        ownProps.history.push('/confirmation-sent');
      })
      .catch((err) => console.error(err));
    },
    onRegisterOneTimePasswordClicked: () => {
      fetch('/api/secondfactor/totp/identity/start', {
        method: 'POST',
      })
      .then(async (res) => {
        if (res.status != 200) {
          throw new Error('Status code ' + res.status);
        }
        ownProps.history.push('/confirmation-sent');
      })
      .catch((err) => console.error(err));
    },
    onStateLoaded: async (state: RemoteState) => {
      if (state.authentication_level < AuthenticationLevel.ONE_FACTOR) {
        ownProps.history.push('/');
        return;
      }
      await triggerSecurityKeySigning();
    }
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(SecondFactorView);