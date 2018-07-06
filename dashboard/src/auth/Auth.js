import React, { Component } from 'react';
import api from '../api/api.js';
import Login from './Login.js';
import LogoutContext from './LogoutContext.js';
import { getSessionItem, setSessionItem } from "../util/Storage.js";

const tokenSessionKey = "auth-token";

class Auth extends Component {
  state = {
    authenticated: false,
    token: getSessionItem(tokenSessionKey) || ""
  };

  async componentDidMount() {
    try {
      api.token = this.state.token;
      await api.get('/api/operators');
      this.setState({
        authenticated: true,
        token: api.token
      });
    } catch (e) {
      this.setState({
        authenticated: false,
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
        {(!this.state.authenticated) ? 
          <Login doLogin={this.handleLogin} error={this.state.error}/> :
          this.props.children
        }
      </LogoutContext.Provider>
    );
  }
}

export default Auth;
