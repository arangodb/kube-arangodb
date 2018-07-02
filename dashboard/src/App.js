import React, { Component } from 'react';
import DeploymentOperator from './deployment/DeploymentOperator.js';
import NoOperator from './NoOperator.js';
import { apiGet } from './api/api.js';
import { Container, Segment, Message } from 'semantic-ui-react';
import './App.css';

class App extends Component {
  constructor() {
    super();
    this.state = {operators:{}};
  }

  async componentDidMount() {
    this.reloadOperators();
  }

  async reloadOperators() {
    const operators = await apiGet('/api/operators');
    this.setState({operators});
  }

  render() {
    setTimeout(this.reloadOperators.bind(this), 5000);
    let children = [];
    if (this.state.operators.deployment) {
      children.push((<DeploymentOperator/>));
    } else {
      children.push((<NoOperator/>));
    }
    return (
      <Container>
        {children}
        <Segment basic>
          <Message>
            <Message.Header>Kube-ArangoDB</Message.Header>
            <p>
              Running in Pod <b>{this.state.operators.pod}</b> in namespace <b>{this.state.operators.namespace}</b>.
            </p>
          </Message>     
        </Segment>
      </Container>
    );
  }
}

export default App;
