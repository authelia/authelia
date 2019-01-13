import { connect } from 'react-redux';
import StateSynchronizer, { OnLoaded, OnError } from '../../../components/StateSynchronizer/StateSynchronizer';
import { RootState } from '../../../reducers';
import { fetchStateSuccess, fetchState, fetchStateFailure } from '../../../reducers/Portal/actions';
import RemoteState from '../../../reducers/Portal/RemoteState';
import { Dispatch } from 'redux';

const mapStateToProps = (state: RootState) => ({
  state: state.remoteState,
  stateError: state.remoteStateError,
  stateLoading: state.remoteStateLoading,
});

const mapDispatchToProps = (dispatch: Dispatch) => {
  return {
    fetch: (onloaded: OnLoaded, onerror: OnError) => {
      dispatch(fetchState());
      fetch('/api/state').then(async (res) => {
        const body = await res.json() as RemoteState;
        await dispatch(fetchStateSuccess(body));
        await onloaded(body);
      })
      .catch(async (err) => {
        await dispatch(fetchStateFailure(err));
        await onerror(err);
      })
    }
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(StateSynchronizer);