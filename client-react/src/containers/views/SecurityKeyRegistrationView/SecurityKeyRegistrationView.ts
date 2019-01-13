import { connect } from 'react-redux';
import SecurityKeyRegistrationView from '../../../views/SecurityKeyRegistrationView/SecurityKeyRegistrationView';
import { RootState } from '../../../reducers';
import { Dispatch } from 'redux';
import {to} from 'await-to-js';
import * as U2fApi from "u2f-api";

const mapStateToProps = (state: RootState) => ({});

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

const mapDispatchToProps = (dispatch: Dispatch) => {
  return {
    componentDidMount: async (token: string) => {
      let err, result;
      [err, result] = await to(checkIdentity(token));
      if (err) {
        console.error(err);
        return;
      }
      [err, result] = await to(requestRegistration());
      if (err) {
        console.error(err);
        return;
      }

      [err, result] = await to(U2fApi.register(result, [], 60));
      if (err) {
        console.error(err);
        return;
      }

      [err, result] = await to(completeRegistration(result as U2fApi.RegisterResponse));
      if (err) {
        console.error(err);
        return;
      }
    }
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(SecurityKeyRegistrationView);