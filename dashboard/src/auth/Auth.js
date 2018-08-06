import React, { Component } from 'react';

import { getSessionItem, setSessionItem } from '../util/Storage';
import api from '../api/api';
import Loading from '../util/Loading';
import Login from './Login';
import LogoutContext from './LogoutContext';

const tokenSessionKey = "auth-token";

// withAuth adds a doLogout property to the given component.
export function withAuth(WrappedComponent) {
  return function AuthAwareComponent(props) {
      return (
        <LogoutContext.Consumer>
          {doLogout => <WrappedComponent {...props} doLogout={doLogout} />}
        </LogoutContext.Consumer>
      );
  }
}

class Auth extends Component {
  state = {
    authenticated: false,
    showLoading: true,
    token: getSessionItem(tokenSessionKey) || ""
  };

  async componentDidMount() {
    try {
      api.token = this.state.token;
      await api.get('/api/operators');
      this.setState({
        authenticated: true,
        showLoading: false,
        token: api.token
      });
    } catch (e) {
      this.setState({
        authenticated: false,
        showLoading: false,
        token: ''
      });
    }
  }

  handleLogin = async (username, password) => {
    try {
      this.setState({error:undefined});
      const res = await api.post('/login', { username, password });
      api.token = res.token;
      setSessionItem(tokenSessionKey, res.token);
      this.setState({
        authenticated: true,
        token: res.token
      });
      return true;
    } catch (e) {
      this.setState({
        authenticated: false,
        token: '',
        error: e.message
      });
      return false;
    }
  };

  handleLogout = () => {
    api.token = '';
    setSessionItem(tokenSessionKey, '');
    this.setState({
      authenticated: false,
      token: '',
      error: undefined
    });
  };

  componentWillUnmount() {
  }

  render() {
    return (
      <LogoutContext.Provider value={this.handleLogout}>
        {(this.state.showLoading) ? <Loading/> : 
           (!this.state.authenticated) ? 
            <Login doLogin={this.handleLogin} error={this.state.error}/> :
            this.props.children
        }
      </LogoutContext.Provider>
    );
  }
}

export default Auth;
