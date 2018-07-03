import React, { Component } from 'react';
import DeploymentOperator from './deployment/DeploymentOperator.js';
import NoOperator from './NoOperator.js';
import Loading from './util/Loading.js';
import { apiGet } from './api/api.js';
import { Container, Segment, Message } from 'semantic-ui-react';
import './App.css';

const OperatorsView = ({deployment, pod, namespace}) => (
  <Container>
    {deployment ? <DeploymentOperator /> : <NoOperator />}
    <Segment basic>
      <Message>
        <Message.Header>Kube-ArangoDB</Message.Header>
        <p>Running in Pod <b>{pod}</b> in namespace <b>{namespace}</b>.</p>
      </Message>
    </Segment>
  </Container>
);

const LoadingView = () => (
  <Container>
    <Loading/>
  </Container>
);

class App extends Component {
  state = {};

  componentDidMount() {
    this.intervalId = setInterval(this.reloadOperators, 5000);
    this.reloadOperators();
  }

  componentWillUnmount() {
    clearInterval(this.intervalId);
  }

  reloadOperators = async() => {
    const operators = await apiGet('/api/operators');
    this.setState({operators});
  }

  render() {
    if (this.state.operators) {
      return <OperatorsView 
        deployment={this.state.operators.deployment} 
        pod={this.state.operators.pod} 
        namespace={this.state.operators.namespace} 
      />;
    }
    return (<LoadingView/>);
  }
}

export default App;
