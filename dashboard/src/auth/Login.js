import { Button, Container, Form, Icon, Message, Modal } from 'semantic-ui-react';
import { css } from 'react-emotion';
import React, { Component } from 'react';

const LoginView = ({username, password, onUsernameChanged, onPasswordChanged, doLogin, error}) => (
  <Container>
    <Form onSubmit={doLogin}>
      <Form.Field>
        <label>Name</label>
        <input focus="true" value={username} onChange={(e) => onUsernameChanged(e.target.value)}/>
      </Form.Field>
      <Form.Field>
        <label>Password</label>
        <input type="password" value={password} onChange={(e) => onPasswordChanged(e.target.value)}/>
      </Form.Field>
      <Form.Button className={css`display:none`} type="submit" />
    </Form>
    {(error) ? <Message error content={error}/> : null}
  </Container>
);

class Login extends Component {
  state = {
    username: '',
    password: ''
  };

  handleLogin = () => {
    this.props.doLogin(this.state.username, this.state.password);
  }

  render() {
    return (
      <Modal open>
        <Modal.Header>Login</Modal.Header>
        <Modal.Content>
          <LoginView 
            error={this.props.error}
            username={this.state.username}
            password={this.state.password}
            onUsernameChanged={(v) => this.setState({username:v})}
            onPasswordChanged={(v) => this.setState({password:v})}
            doLogin={this.handleLogin}
          />
        </Modal.Content>
        <Modal.Actions>
          <Button color='green' disabled={((!this.state.username) || (!this.state.password))} onClick={this.handleLogin}>
            <Icon name='checkmark' /> Login
          </Button>
        </Modal.Actions>
      </Modal>
    );
  }
}

export default Login;
